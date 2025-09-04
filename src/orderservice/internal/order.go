package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/mail"
	"regexp"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	storage  OrderStorage
	rabbitmq RabbitMQ
	clients  ServiceClients
}

// external service clients
type ServiceClients struct {
	Coupon   pb.CouponServiceClient
	Customer pb.CustomerServiceClient
	Delivery pb.DeliveryServiceClient
	Merchant pb.MerchantServiceClient
}

func NewOrderService(s OrderStorage, rb RabbitMQ, cl ServiceClients) *OrderService {
	return &OrderService{storage: s, rabbitmq: rb, clients: cl}
}

func (x *OrderService) ListOrderHistory(ctx context.Context,
	in *pb.ListOrderHistoryRequest) (*pb.ListOrderHistoryResponse, error) {

	if in.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "ID must be provided")
	}

	dbOrders, err := x.storage.ListPlaceOrders(ctx, in.CustomerId)
	if err != nil {
		slog.Error("storage list place orders", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var placeOrders []*pb.PlaceOrder
	for _, dbOrder := range dbOrders {
		placeOrder := toProtoPlaceOrder(dbOrder)
		placeOrders = append(placeOrders, placeOrder)
	}

	return &pb.ListOrderHistoryResponse{PlaceOrders: placeOrders}, nil
}

func (x *OrderService) CreatePlaceOrder(ctx context.Context, in *pb.CreatePlaceOrderRequest) (*pb.PlaceOrder, error) {

	newOrder, err := x.prepareNewOrder(in)
	if err != nil {
		// TODO: include prepare error in response(customer errors).
		return nil, status.Error(codes.FailedPrecondition, "failed to prepare new order")
	}

	orderID, err := x.storage.Create(ctx, newOrder)
	if err != nil {
		slog.Error("storage create new order", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	dbOrder, err := x.storage.GetPlaceOrder(ctx, orderID)
	if err != nil {
		slog.Error("storage get place order", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	order := toProtoPlaceOrder(dbOrder)

	body, err := proto.Marshal(order)
	if err != nil {
		slog.Error("protobuf marshal failed", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	err = x.rabbitmq.Publish(ctx, "order.placed.event", amqp.Publishing{
		Type: "ihavefood.PlaceOrder",
		Body: body,
	})
	if err != nil {
		slog.Error("rabbitmq publish", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	slog.Info("published event", "orderId", order.OrderId)

	return order, nil
}

//---------------------------------------------------------------------------------------

func (x *OrderService) StartConsume() {

	registerEvent := []struct {
		queue   string
		key     string
		handler chan<- amqp.Delivery
	}{
		{"order_status_update_queue", "rider.finding.event", x.handleOrderStatus()},
		{"order_status_update_queue", "merchant.accepted.event", x.handleOrderStatus()},
		{"order_status_update_queue", "rider.assigned.event", x.handleOrderStatus()},
		{"order_status_update_queue", "rider.delivered.event", x.handleOrderStatus()},
		{"payment_status_update_queue", "order.paid.event", x.handlePaymentStatus()},
	}

	for _, r := range registerEvent {
		go func(queue, key string, handler chan<- amqp.Delivery) {

			deliveries, err := x.rabbitmq.Subscribe(context.TODO(), queue, key)
			if err != nil {
				slog.Error("subscribe order", "err", err)
			}

			for delivery := range deliveries {
				handler <- delivery
			}

		}(r.queue, r.key, r.handler)
	}

	select {}
}

func (x *OrderService) handleOrderStatus() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)
	var status pb.OrderStatus

	go func() {
		for msg := range messages {

			switch msg.RoutingKey {
			case "rider.finding.event":
				status = pb.OrderStatus_FINDING_RIDER
			case "merchant.accepted.event":
				status = pb.OrderStatus_PREPARING_ORDER
			case "food.ready.event":
				status = pb.OrderStatus_WAIT_FOR_PICKUP
			case "rider.assigned.event":
				status = pb.OrderStatus_ONGOING
			case "rider.delivered.event":
				status = pb.OrderStatus_DELIVERED
			default:
				slog.Error("unknown routing key %s", "key", msg.RoutingKey)
				continue
			}

			var orderID string
			if err := json.Unmarshal(msg.Body, &orderID); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdateOrderStatus(context.TODO(),
				orderID, dbOrderStatus(status)); err != nil {
				slog.Error("updated status", "err", err, "orderId", orderID)
				continue
			}
		}
	}()
	return messages
}

// handlePaymentStatus updates payment status of an order after
// it has been successfully processed.
//   - Cash method update when rider received cash from customer after delivery.
//   - PromptPay and Credit card upon succussful transaction.
func (x *OrderService) handlePaymentStatus() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	go func() {
		for msg := range messages {

			if msg.RoutingKey != "customer.paid.event" {
				slog.Error("unknown routing key ", "key", msg.RoutingKey)
				continue
			}

			var orderID string
			if err := json.Unmarshal(msg.Body, &orderID); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdatePaymentStatus(
				context.TODO(),
				orderID,
				PaymentStatus_PAID,
			); err != nil {
				slog.Error("updated status", "err", err, "orderId", orderID)
				continue
			}

			// TODO publish update
		}
	}()
	return messages
}

// - add context
func (x *OrderService) prepareNewOrder(newOrder *pb.CreatePlaceOrderRequest) (*newPlaceOrder, error) {
	ctx := context.TODO()

	if err := validatePlaceOrderRequest(newOrder); err != nil {
		return nil, err
	}

	// TODO: call check customer exists instead
	_, err := x.clients.Customer.GetCustomer(ctx, &pb.GetCustomerRequest{
		CustomerId: newOrder.CustomerId,
	})
	if err != nil {
		return nil, err
	}

	deliveryFee, err := x.clients.Delivery.GetDeliveryFee(ctx, &pb.GetDeliveryFeeRequest{
		CustomerId:        newOrder.CustomerId,
		CustomerAddressId: newOrder.CustomerAddress.AddressId,
		MerchantId:        newOrder.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	merchant, err := x.clients.Merchant.GetMerchant(ctx, &pb.GetMerchantRequest{MerchantId: newOrder.MerchantId})
	if err != nil {
		return nil, err
	}

	foodCost := calcFoodCost(merchant.Menu, newOrder.Items)

	coupon, err := x.clients.Coupon.GetCoupon(ctx, &pb.GetCouponRequest{Code: newOrder.CouponCode})
	if err != nil {
		return nil, err
	}

	if time.Now().After(time.Unix(coupon.ExpiresIn, 0)) || coupon.QuantityCount <= 0 {
		return nil, errors.New("coupon is expired or has no remaining uses")
	}

	total := (deliveryFee.Fee + foodCost) - coupon.Discount

	var items []*dbOrderItem
	for _, item := range newOrder.Items {
		items = append(items, &dbOrderItem{
			ItemID:   item.ItemId,
			Quantity: item.Quantity,
			Note:     item.Note,
		})
	}

	return &newPlaceOrder{
		RequestID:       newOrder.RequestId,
		CustomerID:      newOrder.CustomerId,
		MerchantID:      newOrder.MerchantId,
		Items:           items,
		CouponCode:      newOrder.CouponCode,
		CouponDiscount:  coupon.Discount,
		DeliveryFee:     deliveryFee.Fee,
		Total:           total,
		CustomerAddress: toDbAddress(newOrder.CustomerAddress),
		MerchantAddress: toDbAddress(merchant.Address),
		CustomerContact: toDbContactInfo(newOrder.CustomerContact),
		PaymentMethods:  dbPaymentMethods(newOrder.PaymentMethods),
	}, nil
}

func calcFoodCost(menu []*pb.MenuItem, orderItems []*pb.OrderItem) int32 {
	menuPrices := make(map[string]int32)
	for _, menuItem := range menu {
		menuPrices[menuItem.ItemId] = menuItem.Price
	}

	var foodCost int32
	for _, newItem := range orderItems {
		if price, ok := menuPrices[newItem.ItemId]; ok {
			foodCost += price * newItem.Quantity
		}
	}
	return foodCost
}

// TODO: use validator lib
func validatePlaceOrderRequest(in *pb.CreatePlaceOrderRequest) error {

	if in.RequestId == "" {
		return errors.New("request ID must be provided")
	}
	if in.CustomerId == "" {
		return errors.New("customer ID must be provided")
	}
	if in.MerchantId == "" {
		return errors.New("merchant ID must be provided")
	}
	if len(in.Items) == 0 {
		return errors.New("items should be at least one")
	}
	if in.CouponCode == "" {
		return errors.New("coupon code must be provided")
	}
	if in.CustomerAddress == nil {
		return errors.New("customer address must be provided")
	}
	if in.CustomerContact == nil {
		return errors.New("customer contact information must be provided")
	}
	if in.PaymentMethods == pb.PaymentMethods_PAYMENT_METHOD_UNSPECIFIED {
		return errors.New("payment method must be provided")
	}
	if _, ok := pb.PaymentStatus_name[int32(in.PaymentMethods)]; !ok {
		return errors.New("invalid payment method")
	}

	if err := validateEmail(in.CustomerContact.Email); err != nil {
		return err
	}

	if err := validatePhoneNumber(in.CustomerContact.PhoneNumber); err != nil {
		return err
	}

	return nil
}

// validatePhoneNumber validates a customer's phone number according to the Thailand
// phone number format (e.g., 06XXXXXXXX, 08XXXXXXXX, 09XXXXXXXX).
// Any format outside of this is considered invalid, and the function returns an error.
func validatePhoneNumber(phoneNumber string) error {
	if !regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber) {
		return errors.New("invalid phone number format")
	}
	return nil
}

// validateEmail validates the customer's email address to ensure it follows
// the standard email format. It uses mail.ParseAddress to parse the email.
// If the email is invalid, it returns an error.
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

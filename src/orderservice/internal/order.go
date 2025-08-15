package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	storage  OrderStorage
	rabbitmq RabbitMQ
}

func NewOrderService(storage OrderStorage, rabbitmq RabbitMQ) *OrderService {
	return &OrderService{
		storage:  storage,
		rabbitmq: rabbitmq,
	}
}

func (x *OrderService) ListOrderHistory(ctx context.Context,
	in *pb.ListOrderHistoryRequest) (*pb.ListOrderHistoryResponse, error) {

	if in.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "ID must be provided")
	}

	dbOrders, err := x.storage.ListPlaceOrders(ctx, in.CustomerId)
	if err != nil {
		slog.Error("failed to list place orders", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrieve customer's place orders")
	}

	var placeOrders []*pb.PlaceOrder
	for _, dbOrder := range dbOrders {
		placeOrder := dbToProto(dbOrder)
		placeOrders = append(placeOrders, placeOrder)
	}

	return &pb.ListOrderHistoryResponse{PlaceOrders: placeOrders}, nil
}

// HandlePlaceOrder processes an incoming order placement request from the client.
//
// This function validates the place order request, saves the order details to the database,
// and publishes an "order.placed.event" to other services for further processing.
func (x *OrderService) HandlePlaceOrder(ctx context.Context,
	in *pb.HandlePlaceOrderRequest) (*pb.PlaceOrder, error) {

	if err := validatePlaceOrderRequest(in); err != nil {
		slog.Error("validation failed", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate place order request")
	}

	// TODO validate place order fields valid such as customerID , restuarntName already exists

	orderID, err := x.storage.Create(ctx, prepareNewOrder(in))
	if err != nil {
		slog.Error("failed to insert place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to save place order")
	}

	dbOrder, err := x.storage.GetPlaceOrder(ctx, orderID)
	if err != nil {
		slog.Error("failed to retrive place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrive place order")
	}

	order := dbToProto(dbOrder)

	body, err := proto.Marshal(order)
	if err != nil {
		slog.Error("failed to marshal order", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to marshal")
	}

	err = x.rabbitmq.Publish(ctx, "order.placed.event", amqp.Publishing{
		Type: "ihavefood.PlaceOrder",
		Body: body,
	})
	if err != nil {
		slog.Error("failed to publish an event", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to publish event")
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

func prepareNewOrder(in *pb.HandlePlaceOrderRequest) *dbPlaceOrder {

	var menu []*dbMenuItem
	for _, m := range in.Menu {
		menu = append(menu, &dbMenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	createTime := time.Now()

	return &dbPlaceOrder{
		//OrderID:  - ,
		RequestID:      in.RequestId,
		CustomerID:     in.CustomerId,
		MerchantID:     in.MerchantId,
		Menu:           menu,
		CouponCode:     in.CouponCode,
		CouponDiscount: in.CouponDiscount,
		DeliveryFee:    in.DeliveryFee,
		Total:          in.Total,
		CustomerAddress: &dbAddress{
			AddressName: in.CustomerAddress.AddressName,
			SubDistrict: in.CustomerAddress.SubDistrict,
			District:    in.CustomerAddress.District,
			Province:    in.CustomerAddress.Province,
			PostalCode:  in.CustomerAddress.Province,
		},
		MerchantAddress: &dbAddress{
			AddressName: in.CustomerAddress.AddressName,
			SubDistrict: in.CustomerAddress.SubDistrict,
			District:    in.CustomerAddress.District,
			Province:    in.CustomerAddress.Province,
			PostalCode:  in.CustomerAddress.Province,
		},
		CustomerContact: &dbContactInfo{
			PhoneNumber: in.CustomerContact.PhoneNumber,
			Email:       in.CustomerContact.Email,
		},
		PaymentMethods: dbPaymentMethods(in.PaymentMethods),
		PaymentStatus:  PaymentStatus_UNPAID, //DEFAULT
		OrderStatus:    OrderStatus_PENDING,  //DEFAULT
		Timestamps: &dbTimestamps{
			CreateTime:   createTime,
			UpdateTime:   time.Time{},
			CompleteTime: time.Time{},
		}}
}

func dbToProto(order *dbPlaceOrder) *pb.PlaceOrder {

	var menu []*pb.MenuItem
	for _, m := range order.Menu {
		menu = append(menu, &pb.MenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})

	}

	return &pb.PlaceOrder{
		OrderId:        order.OrderID,
		RequestId:      order.RequestID,
		CustomerId:     order.CustomerID,
		MerchantId:     order.MerchantID,
		Menu:           menu,
		CouponCode:     order.CouponCode,
		CouponDiscount: order.CouponDiscount,
		DeliveryFee:    order.DeliveryFee,
		Total:          order.Total,
		CustomerAddress: &pb.Address{
			AddressName: order.CustomerAddress.AddressName,
			SubDistrict: order.CustomerAddress.SubDistrict,
			District:    order.CustomerAddress.District,
			Province:    order.CustomerAddress.Province,
			PostalCode:  order.CustomerAddress.PostalCode,
		},
		MerchantAddress: &pb.Address{
			AddressName: order.MerchantAddress.AddressName,
			SubDistrict: order.MerchantAddress.SubDistrict,
			District:    order.MerchantAddress.District,
			Province:    order.MerchantAddress.Province,
			PostalCode:  order.MerchantAddress.PostalCode,
		},
		CustomerContact: &pb.ContactInfo{
			PhoneNumber: order.CustomerContact.PhoneNumber,
			Email:       order.CustomerContact.Email,
		},
		PaymentMethods: pb.PaymentMethods(order.PaymentMethods),
		PaymentStatus:  pb.PaymentStatus(order.PaymentStatus),
		OrderStatus:    pb.OrderStatus(order.OrderStatus),
		OrderTimestamps: &pb.OrderTimestamps{
			CreateTime:   timestamppb.New(order.Timestamps.CreateTime),
			UpdateTime:   timestamppb.New(order.Timestamps.UpdateTime),
			CompleteTime: timestamppb.New(order.Timestamps.CompleteTime),
		},
	}

}

// TODO implement other validation technique
func validatePlaceOrderRequest(in *pb.HandlePlaceOrderRequest) error {

	switch {
	case in.RequestId == "":
		return errors.New("request ID must be provided")
	case in.CustomerId == "":
		return errors.New("customer ID must be provided")
	case in.MerchantId == "":
		return errors.New("merchant ID must be provided")
	case len(in.Menu) == 0:
		return errors.New("menu should be at least one")
	case in.CouponCode == "":
		return errors.New("coupon code must be provided")
	case in.DeliveryFee == 0:
		return errors.New("delivery fee should not be zero")
	case in.Total == 0:
		return errors.New("total should not be zero")
	case in.CustomerAddress == nil:
		return errors.New("customer address must be provided")
	case in.MerchantAddress == nil:
		return errors.New("merchant address must be provided")
	case in.CustomerContact == nil:
		return errors.New("customer contact infomation must be provided")
	}

	if err := validateEmail(in.CustomerContact.Email); err != nil {
		return err
	}

	if err := validatePhoneNumber(in.CustomerContact.PhoneNumber); err != nil {
		return err
	}

	var sumMenus int32
	for _, menu := range in.Menu {
		sumMenus += menu.Price
	}
	sum := ((sumMenus + in.DeliveryFee) - in.CouponDiscount)

	if in.Total != sum {
		return fmt.Errorf("total mismatch: calculated %d but got %d", sum, in.Total)
	}

	switch in.PaymentMethods {
	case pb.PaymentMethods_PAYMENT_METHOD_CASH,
		pb.PaymentMethods_PAYMENT_METHOD_CREDIT_CARD:
	default:
		return errors.New("invalid payment methods")
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

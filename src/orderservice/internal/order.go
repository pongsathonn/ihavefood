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

func (x *OrderService) ListUserPlaceOrder(ctx context.Context, in *pb.ListUserPlaceOrderRequest) (*pb.ListUserPlaceOrderResponse, error) {

	if in.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username must be provided")
	}

	dbOrders, err := x.storage.PlaceOrders(ctx, in.Username)
	if err != nil {
		slog.Error("retrive place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user's place orders")
	}

	var placeOrders []*pb.PlaceOrder
	for _, dbOrder := range dbOrders {
		placeOrder := dbToProto(dbOrder)
		placeOrders = append(placeOrders, placeOrder)
	}

	return &pb.ListUserPlaceOrderResponse{PlaceOrders: placeOrders}, nil
}

// HandlePlaceOrder processes an incoming order placement request from the client.
//
// This function validates the place order request, saves the order details to the database,
// and publishes an "order.placed.event" to other services for further processing.
func (x *OrderService) HandlePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*pb.PlaceOrder, error) {

	if err := validatePlaceOrderRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate place order request: %v", err)
	}

	orderNO, err := x.storage.Create(ctx, prepareNewOrder(in))
	if err != nil {
		slog.Error("failed to insert place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to save place order")
	}

	dbOrder, err := x.storage.PlaceOrder(ctx, orderNO)
	if err != nil {
		slog.Error("failed to retrive place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrive place order")
	}

	order := dbToProto(dbOrder)

	body, err := proto.Marshal(order)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal: %v", err)
	}

	err = x.rabbitmq.Publish(ctx, "order.placed.event", amqp.Publishing{
		Type: "ihavefood.PlaceOrder",
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish event: %v", err)
	}

	slog.Info("published event", "orderNo", order.OrderNo)

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
		{"order_status_update_queue", "restaurant.accepted.event", x.handleOrderStatus()},
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
			case "restaurant.accepted.event":
				status = pb.OrderStatus_PREPARING_ORDER
			case "food.ready.event":
				status = pb.OrderStatus_WAIT_FOR_PICKUP
			case "rider.assigned.event":
				status = pb.OrderStatus_ONGOING
			case "rider.delivered.event":
				status = pb.OrderStatus_DELIVERED
			default:
				slog.Error("unknown routing key %s", msg.RoutingKey)
				continue
			}

			var orderNO string
			if err := json.Unmarshal(msg.Body, &orderNO); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdateOrderStatus(context.TODO(), orderNO, dbOrderStatus(status)); err != nil {
				slog.Error("updated status", "err", err, "orderNo", orderNO)
				continue
			}
		}
	}()
	return messages
}

// handlePaymentStatus updates payment status of an order after
// it has been successfully processed.
//   - Cash method update when rider received cash from user after delivery.
//   - PromptPay and Credit card upon succussful transaction.
func (x *OrderService) handlePaymentStatus() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	go func() {
		for msg := range messages {

			if msg.RoutingKey != "user.paid.event" {
				slog.Error("unknown routing key %s", msg.RoutingKey)
				continue
			}

			var orderNO string
			if err := json.Unmarshal(msg.Body, &orderNO); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdatePaymentStatus(
				context.TODO(),
				orderNO,
				PaymentStatus_PAID,
			); err != nil {
				slog.Error("updated status", "err", err, "orderNo", orderNO)
				continue
			}

			// TODO publish update
		}
	}()
	return messages
}

func prepareNewOrder(in *pb.HandlePlaceOrderRequest) *dbPlaceOrder {

	var menus []*dbMenu
	for _, m := range in.Menus {
		menus = append(menus, &dbMenu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	createTime := time.Now()

	return &dbPlaceOrder{
		//OrderNo:  - ,
		RequestID:      in.RequestId,
		Username:       in.Username,
		RestaurantNo:   in.RestaurantNo,
		Menus:          menus,
		CouponCode:     in.CouponCode,
		CouponDiscount: in.CouponDiscount,
		DeliveryFee:    in.DeliveryFee,
		Total:          in.Total,
		UserAddress: &dbAddress{
			AddressName: in.UserAddress.AddressName,
			SubDistrict: in.UserAddress.SubDistrict,
			District:    in.UserAddress.District,
			Province:    in.UserAddress.Province,
			PostalCode:  in.UserAddress.Province,
		},
		RestaurantAddress: &dbAddress{
			AddressName: in.UserAddress.AddressName,
			SubDistrict: in.UserAddress.SubDistrict,
			District:    in.UserAddress.District,
			Province:    in.UserAddress.Province,
			PostalCode:  in.UserAddress.Province,
		},
		UserContact: &dbContactInfo{
			PhoneNumber: in.UserContact.PhoneNumber,
			Email:       in.UserContact.Email,
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

	var menus []*pb.Menu
	for _, m := range order.Menus {
		menus = append(menus, &pb.Menu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})

	}

	return &pb.PlaceOrder{
		OrderNo:        order.OrderNo.Hex(),
		RequestId:      order.RequestID,
		Username:       order.Username,
		RestaurantNo:   order.RestaurantNo,
		Menus:          menus,
		CouponCode:     order.CouponCode,
		CouponDiscount: order.CouponDiscount,
		DeliveryFee:    order.DeliveryFee,
		Total:          order.Total,
		UserAddress: &pb.Address{
			AddressName: order.UserAddress.AddressName,
			SubDistrict: order.UserAddress.SubDistrict,
			District:    order.UserAddress.District,
			Province:    order.UserAddress.Province,
			PostalCode:  order.UserAddress.PostalCode,
		},
		RestaurantAddress: &pb.Address{
			AddressName: order.RestaurantAddress.AddressName,
			SubDistrict: order.RestaurantAddress.SubDistrict,
			District:    order.RestaurantAddress.District,
			Province:    order.RestaurantAddress.Province,
			PostalCode:  order.RestaurantAddress.PostalCode,
		},
		UserContact: &pb.ContactInfo{
			PhoneNumber: order.UserContact.PhoneNumber,
			Email:       order.UserContact.Email,
		},
		PaymentMethods: pb.PaymentMethods(order.PaymentMethods),
		PaymentStatus:  pb.PaymentStatus(order.PaymentStatus),
		OrderStatus:    pb.OrderStatus(order.OrderStatus),
		OrderTimestamps: &pb.OrderTimestamps{
			CreateTime:   order.Timestamps.CreateTime.Unix(),
			UpdateTime:   order.Timestamps.UpdateTime.Unix(),
			CompleteTime: order.Timestamps.CompleteTime.Unix(),
		},
	}

}

// TODO implement other validation technique
func validatePlaceOrderRequest(in *pb.HandlePlaceOrderRequest) error {

	switch {
	case in.Username == "":
		return errors.New("username must be provided")
	case in.RestaurantNo == "":
		return errors.New("restaurant number must be provided")
	case len(in.Menus) == 0:
		return errors.New("menu should be at least one")
	case in.CouponCode == "":
		return errors.New("coupon code must be provided")
	case in.DeliveryFee == 0:
		return errors.New("delivery fee should not be zero")
	case in.Total == 0:
		return errors.New("total should not be zero")
	case in.UserAddress == nil:
		return errors.New("user address must be provided")
	case in.RestaurantAddress == nil:
		return errors.New("restaurant address must be provided")
	case in.UserContact == nil:
		return errors.New("user contact infomation must be provided")
	}

	if err := validateEmail(in.UserContact.Email); err != nil {
		return err
	}

	if err := validatePhoneNumber(in.UserContact.PhoneNumber); err != nil {
		return err
	}

	var sumMenus int32
	for _, menu := range in.Menus {
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

// validatePhoneNumber validates a user's phone number according to the Thailand
// phone number format (e.g., 06XXXXXXXX, 08XXXXXXXX, 09XXXXXXXX).
// Any format outside of this is considered invalid, and the function returns an error.
func validatePhoneNumber(phoneNumber string) error {
	if !regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber) {
		return errors.New("invalid phone number format")
	}
	return nil
}

// validateEmail validates the user's email address to ensure it follows
// the standard email format. It uses mail.ParseAddress to parse the email.
// If the email is invalid, it returns an error.
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

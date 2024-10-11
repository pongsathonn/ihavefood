package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

// TODO add doc
const (
	eventOrderCooking = "order.cooking.event"
	eventOrderReady   = "order.ready.event"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	repository OrderRepository
	rabbitmq   RabbitMQ
}

func NewOrderService(repository OrderRepository, rabbitmq RabbitMQ) *OrderService {
	return &OrderService{
		repository: repository,
		rabbitmq:   rabbitmq,
	}
}

func (x *OrderService) ListUserPlaceOrder(ctx context.Context, in *pb.ListUserPlaceOrderRequest) (*pb.ListUserPlaceOrderResponse, error) {

	if in.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username must be provided")
	}

	res, err := x.repository.PlaceOrders(ctx, in.Username)
	if err != nil {
		slog.Error("retrive place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user's place orders")
	}

	if len(res.PlaceOrders) == 0 {
		return nil, status.Error(codes.NotFound, "place order not found")
	}

	return res, nil
}

// HandlePlaceOrder processes an incoming order placement request from the client.
//
// This function validates the place order request, saves the order details to the database,
// and publishes an "order.placed.event" to other services for further processing.
func (x *OrderService) HandlePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*pb.HandlePlaceOrderResponse, error) {

	if err := validatePlaceOrderRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate place order request: %v", err)
	}

	duplicated, err := x.repository.IsDuplicatedOrder(ctx, in)
	if err != nil {
		slog.Error("check order duplicated", "err", err)
		return nil, status.Error(codes.Internal, "failed to check duplicated order")
	}

	if duplicated {
		return nil, status.Error(codes.AlreadyExists, "order duplicated")
	}

	res, err := x.repository.SaveNewPlaceOrder(ctx, in)
	if err != nil {
		slog.Error("save place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to save place order")
	}

	err = x.rabbitmq.Publish(ctx, "order_exchange", "order.placed.event", &pb.PlaceOrder{
		No:                res.OrderNo,
		Username:          in.Username,
		RestaurantNo:      in.RestaurantNo,
		Menus:             in.Menus,
		CouponCode:        in.CouponCode,
		CouponDiscount:    in.CouponDiscount,
		DeliveryFee:       in.DeliveryFee,
		Total:             in.Total,
		UserAddress:       in.UserAddress,
		RestaurantAddress: in.RestaurantAddress,
		UserContact:       in.UserContact,
		PaymentMethods:    in.PaymentMethods,
		PaymentStatus:     res.PaymentStatus,
		OrderStatus:       res.OrderStatus, // "PENDING"
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish event: %v", err)
	}

	slog.Info("published event", "orderNo", res.OrderNo)

	return &pb.HandlePlaceOrderResponse{
		OrderNo:   res.OrderNo,
		CreatedAt: res.Created_at,
	}, nil
}

//---------------------------------------------------------------------------------------

// might move this function to main.go for making overview
// that what application do with topic
func (x *OrderService) RunMessageProcessing() {

	go x.fetch(eventOrderCooking, x.handleOrderStatus(eventOrderCooking))
	go x.fetch(eventOrderReady, x.handleOrderStatus(eventOrderReady))

	/*
		go x.fetch("rider.notified.event", nil)
		go x.fetch("rider.accepted.event", nil)
		go x.fetch("user.paid.event", nil)
	*/

	select {}
}

// TODO handle seperately exchange
func (x *OrderService) fetch(routingKey string, messages chan<- []byte) {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"order_exchange", // exchange
		"",               // queue
		routingKey,       // routing key
	)
	if err != nil {
		slog.Error("subscribe order", "err", err)
	}

	for delivery := range deliveries {
		messages <- delivery.Body
	}
}

func (x *OrderService) handleOrderStatus(topic string) chan<- []byte {

	messages := make(chan []byte)

	go func() {
		for msg := range messages {

			var orderNo string
			if err := json.Unmarshal(msg, &orderNo); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			var newStatus pb.OrderStatus

			switch topic {
			case eventOrderCooking:
				newStatus = pb.OrderStatus_PREPARING_ORDER
			case eventOrderReady:
				newStatus = pb.OrderStatus_WAIT_FOR_PICKUP
			}

			err := x.repository.UpdateOrderStatus(
				context.TODO(),
				orderNo,
				newStatus,
			)
			if err != nil {
				slog.Error("updated order status",
					"err", err,
					"orderNo", orderNo,
					"newStatus", newStatus.String(),
				)
				continue
			}

		}
	}()
	return messages
}

func (x *OrderService) handleTODO() chan<- []byte {
	messages := make(chan []byte)
	go func() {
		for msg := range messages {
			var todo string
			if err := json.Unmarshal(msg, &todo); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}
			// DO SOMETHING after todo event
		}
	}()
	return messages
}

//---------------------------------------------------------------------------------------

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

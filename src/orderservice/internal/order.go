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

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
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

	resp, err := x.repository.PlaceOrders(ctx, in.Username)
	if err != nil {
		slog.Error("retrive place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrieve user's place orders")
	}

	return resp, nil
}

// HandlePlaceOrder processes an incoming order placement request from the client.
//
// This function validates the place order request, saves the order details to the database,
// and publishes an "order.placed.event" to other services for further processing.
func (x *OrderService) HandlePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*pb.HandlePlaceOrderResponse, error) {

	if err := validatePlaceOrderRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate place order request: %v", err)
	}

	res, err := x.repository.SavePlaceOrder(ctx, in)
	if err != nil {
		slog.Error("save place order", "err", err)
		return nil, status.Error(codes.Internal, "failed to save place order")
	}

	err = x.rabbitmq.Publish(ctx, "order_exchange", "order.placed.event", &pb.PlaceOrder{
		OrderId:           res.OrderId,
		OrderTrackingId:   res.OrderTrackingId,
		Username:          in.Username,
		RestaurantId:      in.RestaurantId,
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
		OrderStatus:       res.OrderStatus,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish event: %v", err)
	}

	slog.Info("published event", "order_id", res.OrderId)

	return &pb.HandlePlaceOrderResponse{
		OrderId:         res.OrderId,
		OrderTrackingId: res.OrderTrackingId,
	}, nil
}

func (x *OrderService) consumedDeliveryFeeAndDiscount() (deliveryFee, Discount int32, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	delivery, err := x.rabbitmq.Subscribe(ctx, "order_exchange", "order_validate_queue", "ROUTING KEY HERE")
	if err != nil {
		return 0, 0, err
	}

	var order pb.PlaceOrder

	select {
	case v := <-delivery:
		if err := json.Unmarshal(v.Body, &order); err != nil {
			return 0, 0, err
		}
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	}

	return order.DeliveryFee, order.CouponDiscount, nil

}

func validatePlaceOrderRequest(in *pb.HandlePlaceOrderRequest) error {

	switch {
	case in.Username == "":
		return errors.New("username must be provided")
	case in.RestaurantId == "":
		return errors.New("contact must be provided")
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

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	resp, err := x.repository.PlaceOrder(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user's place orders :%v", err)
	}

	return resp, nil
}

// HandlePlaceOrder handle imcomming place order from client and publish to other services
//
// routing key explanation
//
//   - "order.validate.event" occurs when place order is comming from client
//     DeliveryService will subscribe this and provided delivery fee for and response.
//     CouponService will validate coupon and resonse discount.
//     consumedDeliveryFeeAndDiscount will subscribe to these
//
//   - "order.placed.event" occurs after saved place order into database and get orderId,TrackingId
//     publish to other service delivery find rider etc.
//
// TODO
// - handle error
// - handle duplicated request
func (x *OrderService) HandlePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*pb.HandlePlaceOrderResponse, error) {

	if err := validatePlaceOrderRequest(in); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to validate place order request: %v", err)
	}

	res, err := x.repository.SavePlaceOrder(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("failed to save place order: %v", err)
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
		PaymentMethod:     in.PaymentMethod,
		PaymentStatus:     res.PaymentStatus,
		OrderStatus:       res.OrderStatus,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create event : %v", err)
	}

	log.Printf("published with order id : %s\n", res.OrderId)

	return &pb.HandlePlaceOrderResponse{
		OrderId:         res.OrderId,
		OrderTrackingId: res.OrderTrackingId,
	}, nil
}

// TODO doc
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

	var sumMenus int32
	for _, menu := range in.Menus {
		sumMenus += menu.Price
	}
	sum := ((sumMenus + in.DeliveryFee) - in.CouponDiscount)

	if in.Total != sum {
		return fmt.Errorf("total mismatch: calculated %d but got %d", sum, in.Total)
	}

	if _, ok := pb.PaymentMethod_value[in.PaymentMethod.String()]; !ok {
		return errors.New("payment methods invalid")
	}
	return nil
}

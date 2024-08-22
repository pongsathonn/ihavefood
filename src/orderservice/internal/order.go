package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	repository OrderRepository
	rabbitmq   RabbitmqClient
}

func NewOrderService(repository OrderRepository, rabbitmq RabbitmqClient) *OrderService {
	return &OrderService{
		repository: repository,
		rabbitmq:   rabbitmq,
	}
}

func (x *OrderService) ListUserPlaceOrder(ctx context.Context, in *pb.ListUserPlaceOrderRequest) (*pb.ListUserPlaceOrderResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	resp, err := x.repository.PlaceOrder(in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user's place orders :%v", err)
	}

	return resp, nil

}

func (x *OrderService) PlaceOrder(ctx context.Context, in *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	if in.Username == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "username or address must be provided")
	}

	// validate payment methods
	pm := in.PaymentMethod.String()
	if _, ok := pb.PaymentMethod_value[pm]; !ok {
		return nil, fmt.Errorf("payment methods invalid: %s", pm)
	}

	var total int32
	for _, mn := range in.Menus {
		total += mn.Price
	}

	if in.Total != ((total + in.DeliveryFee) - in.CouponDiscount) {
		return nil, fmt.Errorf("total invalid")
	}

	// save place order
	res, err := x.repository.SavePlaceOrder(in)
	if err != nil {
		return nil, fmt.Errorf("failed to save place order: %v", err)
	}

	p := &pb.PlaceOrder{
		OrderId:         res.OrderId,
		OrderTrackingId: res.OrderTrackingId,
		Username:        in.Username,
		RestaurantName:  in.RestaurantName,
		Menus:           in.Menus,
		CouponCode:      in.CouponCode,
		CouponDiscount:  in.CouponDiscount,
		DeliveryFee:     in.DeliveryFee,
		Total:           in.Total,
		Address:         in.Address,
		Contact:         in.Contact,
		PaymentMethod:   in.PaymentMethod,
		PaymentStatus:   res.PaymentStatus,
		OrderStatus:     res.OrderStatus,
	}

	body, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %v", err)
	}

	// publish event
	err = x.rabbitmq.Publish(ctx, "order_exchange", "order.placed.event", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create event : %v", err)
	}

	log.Printf("published with order id : %s\n", p.OrderId)

	return &pb.PlaceOrderResponse{
		OrderId:         res.OrderId,
		OrderTrackingId: res.OrderTrackingId,
	}, nil
}

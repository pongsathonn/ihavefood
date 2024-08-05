package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/pongsathonn/ihavefood/src/orderservice/rabbitmq"
	"github.com/pongsathonn/ihavefood/src/orderservice/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

type PlaceOrderBody struct {
	OrderId        string
	TrackingId     string
	Username       string
	RestaurantName string
	Menus          []*pb.Menu
	CouponCode     string
	CouponDiscount int32
	DeliveryFee    int32
	Total          int32
	UserAddress    *pb.Address
	ContactInfo    *pb.ContactInfo
	PaymentMethod  pb.PaymentMethod
	PaymentStatus  pb.PaymentStatus
	OrderStatus    pb.OrderStatus
}

func NewOrder(db repository.OrderRepo, ps rabbitmq.RabbitMQ) *order {
	return &order{
		db: db,
		ps: ps,
	}
}

type order struct {
	pb.UnimplementedOrderServiceServer

	db repository.OrderRepo
	ps rabbitmq.RabbitMQ
}

func (or *order) ListUserPlaceOrder(ctx context.Context, in *pb.ListUserPlaceOrderRequest) (*pb.ListUserPlaceOrderResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username shouldn't be empty")
	}

	resp, err := or.db.PlaceOrder(in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return resp, nil

}

func (or *order) PlaceOrder(ctx context.Context, in *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	if in.Username == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request ")
	}

	pm := in.PaymentMethod.String()
	if _, ok := pb.PaymentMethod_value[pm]; !ok {
		return nil, fmt.Errorf("bad request kuy")
	}

	var total int32
	for _, mn := range in.Menus {
		total += mn.Price
	}

	if in.Total != ((total + in.DeliveryFee) - in.CouponDiscount) {
		return nil, errors.New("total invalid")
	}

	// save place order
	res, err := or.db.SavePlaceOrder(in)
	if err != nil {
		return nil, fmt.Errorf("save failed %v", err)
	}

	/* THIS WORK
	body := &PlaceOrderBody{
		OrderId:         res.OrderId,
		TrackingId:      res.OrderTrackingId,
		Username:        in.Username,
		RestaurantName:  in.RestaurantName,
		Menus:           in.Menus,
		CouponCode:      in.CouponCode,
		CouponDiscount:  in.CouponDiscount,
		DeliveryFee:     in.DeliveryFee,
		Total:           in.Total,
		DeliveryAddress: in.Address,
		ContactInfo:     in.Contact,
		PaymentMethod:   in.PaymentMethod,
		PaymentStatus:   res.PaymentStatus,
		OrderStatus:     res.OrderStatus,
	}
	*/

	body := &pb.PlaceOrder{
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

	// publish event
	routingKey := "order.placed.event"
	err = or.ps.Publish(routingKey, body)
	if err != nil {
		return nil, fmt.Errorf("couldn't create event")
	}
	log.Printf("published with order id : %s\n", body.OrderId)

	// response
	return &pb.PlaceOrderResponse{
		OrderId:         res.OrderId,
		OrderTrackingId: res.OrderTrackingId,
	}, nil
}

package main

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
)

type coupon struct {
	pb.UnimplementedCouponServiceServer

	rb          *amqp.Connection
	orderClient pb.OrderServiceClient
}

func NewCoupon(rb *amqp.Connection, oc pb.OrderServiceClient) *coupon {
	return &coupon{rb: rb, orderClient: oc}
}

func (s *coupon) GetCoupon(ctx context.Context, in *pb.GetCouponRequest) (*pb.GetCouponResponse, error) {

	if in.Code != "" {
		return &pb.GetCouponResponse{
			Coupon: &pb.Coupon{
				Types:    pb.CouponType_COUPON_TYPE_FREE_DELIVERY,
				Code:     "xxxx777xx",
				Discount: 50,
				Period:   20,
				Amount:   5,
			},
		}, nil
	}

	return nil, status.Errorf(codes.Unimplemented, "method GetCoupon not implemented")
}

func (s *coupon) ListCoupon(context.Context, *pb.Empty) (*pb.ListCouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCoupon not implemented")
}

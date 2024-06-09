package handler

import (
	"context"

	pb "github.com/pongsathonn/food-delivery/src/coupon/genproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type couponServer struct {
	pb.UnimplementedCouponServiceServer
}

func NewCouponServer() *couponServer {
	return &couponServer{}
}

func Helloworldkub() {
	println("XXX")
}

func (cs *couponServer) GetCoupon(_ context.Context, in *pb.GetCouponRequest) (*pb.Coupon, error) {
	return &pb.Coupon{Types: pb.CouponType_COUPON_TYPE_DISCOUNT, Code: "asodhi10", Period: 5, Amount: 20}, nil
}
func (cs *couponServer) ListCoupon(_ context.Context, in *pb.Empty) (*pb.ListCouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCoupon not implemented")
}

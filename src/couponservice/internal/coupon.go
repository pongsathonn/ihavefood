package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CouponService struct {
	pb.UnimplementedCouponServiceServer

	rabbitmq   *amqp.Connection
	repository CouponRepository
}

func NewCouponService(rb *amqp.Connection, rp CouponRepository) *CouponService {
	return &CouponService{rabbitmq: rb, repository: rp}
}

// TODO improve doc
//
// Addcoupon is a function insert new coupon . if coupon already exists ( same code )
// it will increase coupon's quantity instead and replace expiration time with the longest .
// coupon type FREE_DELIVERY will ignore field discount so everytime input has type FREE_COUPON
// it will assign discount variable to zero
func (x *CouponService) AddCoupon(ctx context.Context, in *pb.AddCouponRequest) (*pb.AddCouponResponse, error) {

	if in.Discount < 1 || in.Discount > 99 {
		return nil, status.Errorf(codes.InvalidArgument, "discount must be between 1 and 99")
	}

	if in.Quantity < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be at least 1")
	}

	var code string
	var discount int32

	switch in.CouponType {
	case pb.CouponType_COUPON_TYPE_DISCOUNT:
		code = fmt.Sprintf("SAVE%dFORYOU", in.Discount)
		discount = in.Discount
	case pb.CouponType_COUPON_TYPE_FREE_DELIVERY:
		code = fmt.Sprintf("FREEDELIVERY")
		discount = 0
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid coupon type")
	}

	expiration := time.Now().Add(time.Duration(in.ExpireInHour) * time.Hour)

	err := x.repository.SaveCoupon(ctx, &Coupon{
		Types:      int32(in.CouponType),
		Code:       code,
		Discount:   discount,
		Expiration: expiration.Unix(),
		Quantity:   in.Quantity,
	})
	if err != nil {
		log.Println("Save coupon failed: ", err)
		return nil, status.Errorf(codes.Internal, "failed to add coupon")
	}

	return &pb.AddCouponResponse{Success: true}, nil
}

// TODO improve doc
//
// this fn might be use for check how many remaining coupon
func (x *CouponService) GetCoupon(ctx context.Context, in *pb.GetCouponRequest) (*pb.GetCouponResponse, error) {

	if in.Code == "" {
		return nil, status.Errorf(codes.InvalidArgument, "code must be provided")
	}

	c, err := x.repository.GetCoupon(ctx, in.Code)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "coupon not found")
		}
		log.Println("Get coupon failed:", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get coupon")
	}

	var types pb.CouponType
	switch c.Types {
	case 0:
		types = pb.CouponType_COUPON_TYPE_UNSPECIFICED
	case 1:
		types = pb.CouponType_COUPON_TYPE_DISCOUNT
	case 2:
		types = pb.CouponType_COUPON_TYPE_FREE_DELIVERY
	}

	return &pb.GetCouponResponse{Coupon: &pb.Coupon{
		Types:      types,
		Code:       c.Code,
		Discount:   c.Discount,
		Expiration: c.Expiration,
		Quantity:   c.Quantity,
	}}, nil
}

func (x *CouponService) ListCoupon(ctx context.Context, in *pb.Empty) (*pb.ListCouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCoupon not implemented")
}

func (x *CouponService) UseCoupon(ctx context.Context, in *pb.UserCouponRequest) (*pb.UserCouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UseCoupon not implemented")
}

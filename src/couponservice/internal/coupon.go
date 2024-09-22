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

var (
	errNoCouponCode = status.Errorf(codes.InvalidArgument, "coupon code must be provided")
)

type CouponService struct {
	pb.UnimplementedCouponServiceServer

	rabbitmq   *amqp.Connection
	repository CouponRepository
}

func NewCouponService(rb *amqp.Connection, rp CouponRepository) *CouponService {
	return &CouponService{rabbitmq: rb, repository: rp}
}

func (x *CouponService) AddCoupon(ctx context.Context, in *pb.AddCouponRequest) (*pb.AddCouponResponse, error) {

	var (
		code     string
		discount int32
	)

	switch in.CouponTypes {
	case pb.CouponTypes_COUPON_TYPE_DISCOUNT:
		if in.Discount < 1 || in.Discount > 99 {
			return nil, status.Errorf(codes.InvalidArgument, "discount must be between 1 and 99")
		}
		code = fmt.Sprintf("SAVE%dFORYOU", in.Discount)
		discount = in.Discount
	case pb.CouponTypes_COUPON_TYPE_FREE_DELIVERY:
		code = fmt.Sprintf("FREEDELIVERY")
		discount = 0
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid coupon type")
	}

	if in.Quantity < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be at least 1")
	}

	expiration := time.Now().Add(time.Duration(in.ExpireInHour) * time.Hour)

	err := x.repository.SaveCoupon(ctx, &Coupon{
		Types:      int32(in.CouponTypes),
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

func (x *CouponService) GetCoupon(ctx context.Context, in *pb.GetCouponRequest) (*pb.GetCouponResponse, error) {

	if in.Code == "" {
		return nil, errNoCouponCode
	}

	coupon, err := x.repository.Coupon(ctx, in.Code)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "coupon not found")
		}
		log.Println("Get coupon failed:", err)
		return nil, status.Error(codes.InvalidArgument, "failed to retrive coupon from database")
	}

	if coupon.Expiration <= time.Now().Unix() {
		return nil, status.Errorf(codes.FailedPrecondition, "coupon has expired")
	}

	if coupon.Quantity < 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "coupon quantity is insufficient")
	}

	return &pb.GetCouponResponse{Coupon: &pb.Coupon{
		Types:      pb.CouponTypes(coupon.Types),
		Code:       coupon.Code,
		Discount:   coupon.Discount,
		Expiration: coupon.Expiration,
		Quantity:   coupon.Quantity,
	}}, nil
}

func (x *CouponService) ListCoupon(ctx context.Context, empty *pb.Empty) (*pb.ListCouponResponse, error) {

	listCoupons, err := x.repository.Coupons(ctx)
	if err != nil {
		log.Println("List coupons failed: ", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive list coupons")
	}

	var coupons []*pb.Coupon
	for _, c := range listCoupons {
		coupon := &pb.Coupon{
			Types:      pb.CouponTypes(c.Types),
			Code:       c.Code,
			Discount:   c.Discount,
			Expiration: c.Expiration,
			Quantity:   c.Quantity,
		}
		coupons = append(coupons, coupon)
	}
	return &pb.ListCouponResponse{Coupons: coupons}, nil
}

func (x *CouponService) AppliedCoupon(ctx context.Context, in *pb.AppliedCouponRequest) (*pb.AppliedCouponResponse, error) {

	if in.Code == "" {
		return nil, errNoCouponCode
	}

	if err := x.repository.UpdateCouponQuantity(ctx, in.Code); err != nil {
		log.Println("Update failed", err)
		return nil, status.Errorf(codes.Internal, "failed to update coupon's quantity")
	}

	return &pb.AppliedCouponResponse{Success: true}, nil
}

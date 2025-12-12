package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	errNoCouponCode = status.Error(codes.InvalidArgument, "coupon code must be provided")
)

type CouponService struct {
	pb.UnimplementedCouponServiceServer

	rabbitmq *amqp.Connection
	storage  CouponStorage
}

func NewCouponService(con *amqp.Connection, rp CouponStorage) *CouponService {
	return &CouponService{rabbitmq: con, storage: rp}
}

func (x *CouponService) AddCoupon(ctx context.Context, in *pb.AddCouponRequest) (*pb.Coupon, error) {

	var (
		code     string
		discount int32
	)

	switch in.CouponTypes {
	case pb.CouponTypes_COUPON_TYPE_DISCOUNT:
		if in.Discount < 1 || in.Discount > 99 {
			return nil, status.Error(codes.InvalidArgument, "discount must be between 1 and 99")
		}
		code = fmt.Sprintf("SAVE%d", in.Discount)
		discount = in.Discount
	case pb.CouponTypes_COUPON_TYPE_FREE_DELIVERY:
		code = fmt.Sprintf("FREEDELIVERY")
		discount = 0
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid coupon type")
	}

	if in.Quantity < 1 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be at least 1")
	}

	expiration := time.Now().Add(time.Hour * time.Duration(in.ExpireInHour))

	coupon, err := x.storage.Add(ctx, &dbCoupon{
		Types:      int32(in.CouponTypes),
		Code:       code,
		Discount:   discount,
		Expiration: expiration,
		Quantity:   in.Quantity,
	})
	if err != nil {
		slog.Error("storage add", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.Coupon{
		Types:         pb.CouponTypes(coupon.Types),
		Code:          coupon.Code,
		Discount:      coupon.Discount,
		ExpiresIn:     coupon.Expiration.Unix(),
		QuantityCount: coupon.Quantity,
	}, nil
}

func (x *CouponService) GetCoupon(ctx context.Context, in *pb.GetCouponRequest) (*pb.Coupon, error) {

	if in.Code == "" {
		return nil, errNoCouponCode
	}

	coupon, err := x.storage.GetCoupon(ctx, in.Code)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Error("retrive coupon",
				"code", in.Code,
				"err", err,
			)
			return nil, status.Error(codes.NotFound, "coupon not found")
		}
		slog.Error("retrive coupon", "err", err)
		return nil, status.Error(codes.InvalidArgument, "failed to retrive coupon from database")
	}

	if coupon.Expiration.Before(time.Now()) {
		return nil, status.Error(codes.FailedPrecondition, "coupon has expired")
	}

	if coupon.Quantity < 1 {
		return nil, status.Error(codes.FailedPrecondition, "coupon quantity is insufficient")
	}

	return &pb.Coupon{
		Types:         pb.CouponTypes(coupon.Types),
		Code:          coupon.Code,
		Discount:      coupon.Discount,
		ExpiresIn:     coupon.Expiration.Unix(),
		QuantityCount: coupon.Quantity,
	}, nil
}

func (x *CouponService) ListCoupons(ctx context.Context, empty *emptypb.Empty) (*pb.ListCouponsResponse, error) {

	listCoupons, err := x.storage.ListCoupons(ctx)
	if err != nil {
		slog.Error("storage list coupons", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var coupons []*pb.Coupon
	for _, c := range listCoupons {
		coupon := &pb.Coupon{
			Types:         pb.CouponTypes(c.Types),
			Code:          c.Code,
			Discount:      c.Discount,
			ExpiresIn:     c.Expiration.Unix(),
			QuantityCount: c.Quantity,
		}
		coupons = append(coupons, coupon)
	}
	return &pb.ListCouponsResponse{Coupons: coupons}, nil
}

func (x *CouponService) RedeemCoupon(ctx context.Context, in *pb.RedeemCouponRequest) (*pb.RedeemCouponResponse, error) {

	if in.Code == "" {
		return nil, errNoCouponCode
	}

	if _, err := x.storage.UpdateQuantity(ctx, in.Code); err != nil {
		slog.Error("update coupon quantity", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.RedeemCouponResponse{Success: true}, nil
}

package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Code is the primary identifier
// Types relies on protobuf
// Discount is fixed amount value, not a percentage.
type dbCoupon struct {
	Code     string `bson:"_id"`
	Types    int32  `bson:"types"`
	Discount int32  `bson:"discount"`
	// Expiration must be "UTC"
	Expiration time.Time `bson:"expiration"`
	Quantity   int32     `bson:"quantity"`
}

type CouponStorage interface {
	ListCoupons(ctx context.Context) ([]*dbCoupon, error)

	GetCoupon(ctx context.Context, code string) (*dbCoupon, error)

	// Add inserts new coupon. If the coupon code already exists,
	// update the quantity and expiration time instead
	Add(ctx context.Context, coupon *dbCoupon) (*dbCoupon, error)

	// UpdateQuantity decreases the quantity of the specified coupon
	// by 1.This function is invoked after the coupon has been used.
	UpdateQuantity(ctx context.Context, code string) (string, error)

	// DeleteCoupon(code string) error
}

type couponStorage struct {
	coll *mongo.Collection
}

func NewCouponStorage(coll *mongo.Collection) CouponStorage {
	return &couponStorage{coll: coll}
}

func (s *couponStorage) ListCoupons(ctx context.Context) ([]*dbCoupon, error) {

	// filter := bson.M{"quantity": bson.M{"$gt": 0}}
	cur, err := s.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var coupons []*dbCoupon
	if err := cur.All(ctx, &coupons); err != nil {
		return nil, err
	}

	return coupons, nil

}

func (s *couponStorage) GetCoupon(ctx context.Context, code string) (*dbCoupon, error) {

	var coupon dbCoupon
	if err := s.coll.FindOne(ctx, bson.M{"_id": code}).Decode(&coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (s *couponStorage) Add(ctx context.Context, coupon *dbCoupon) (*dbCoupon, error) {

	// if coupon code exists increase quantity field
	// and update with longest expiration time
	filter := bson.M{"_id": coupon.Code}
	update := bson.M{
		"$setOnInsert": bson.M{
			"types":    coupon.Types,
			"discount": coupon.Discount,
		},
		"$inc": bson.M{"quantity": coupon.Quantity},
		"$max": bson.M{"expiration": coupon.Expiration},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	addedCoupon := &dbCoupon{}
	if err := s.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(addedCoupon); err != nil {
		return nil, err
	}

	return addedCoupon, nil
}

func (s *couponStorage) UpdateQuantity(ctx context.Context, code string) (string, error) {

	filter := bson.M{
		"_id":      code,
		"quantity": bson.M{"$gt": 0},
	}
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "quantity", Value: -1},
		}},
	}

	res, err := s.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		return "", errors.New("coupon code does not exists or quantity is insufficient")
	}

	updatedCode := res.UpsertedID.(string)
	return updatedCode, nil
}

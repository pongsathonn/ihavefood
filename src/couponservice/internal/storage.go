package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// Coupons returns all coupons in database
	Coupons(ctx context.Context) ([]*dbCoupon, error)

	Coupon(ctx context.Context, code string) (*dbCoupon, error)

	// Add inserts new coupon. If the coupon code already exists,
	// update the quantity and expiration time instead
	Add(ctx context.Context, coupon *dbCoupon) (*dbCoupon, error)

	// UpdateQuantity decreases the quantity of the specified coupon
	// by 1.This function is invoked after the coupon has been used.
	UpdateQuantity(ctx context.Context, code string) (string, error)

	// DeleteCoupon(code string) error
}

type couponStorage struct {
	db *mongo.Client
}

func NewCouponStorage(db *mongo.Client) CouponStorage {
	return &couponStorage{db: db}
}

func (r *couponStorage) Coupons(ctx context.Context) ([]*dbCoupon, error) {

	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	// filter := bson.M{"quantity": bson.M{"$gt": 0}}
	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var coupons []*dbCoupon
	if err := cur.All(ctx, &coupons); err != nil {
		return nil, err
	}

	return coupons, nil

}

func (r *couponStorage) Coupon(ctx context.Context, code string) (*dbCoupon, error) {

	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	var coupon dbCoupon
	if err := coll.FindOne(ctx, bson.M{"code": code}).Decode(&coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponStorage) Add(ctx context.Context, coupon *dbCoupon) (*dbCoupon, error) {

	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	// if coupon code exists increase quantity field
	// and update with longest expiration time
	filter := bson.M{"_id": coupon.Code}
	update := bson.M{
		"$inc": bson.M{"quantity": coupon.Quantity},
		"$max": bson.M{"expiration": coupon.Expiration},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var addedCoupon *dbCoupon

	if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(addedCoupon); err != nil {
		return nil, err
	}

	return addedCoupon, nil
}

func (r *couponStorage) UpdateQuantity(ctx context.Context, code string) (string, error) {
	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	filter := bson.M{
		"_id":      code,
		"quantity": bson.M{"$gt": 0},
	}
	update := bson.D{{"$inc", bson.D{{"quantity", -1}}}}

	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		return "", errors.New("coupon code does not exists or quantity is insufficient")
	}

	updatedCode := res.UpsertedID.(string)
	return updatedCode, nil
}

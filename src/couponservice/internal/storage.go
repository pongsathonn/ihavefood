package internal

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Discount is fixed amount value, not a percentage.
type dbCoupon struct {
	Types    int32 //rely on protobuf
	Code     string
	Discount int32

	// TODO time use time.Time instead Unix
	Expiration int64

	Quantity int32
}

type CouponStorage interface {

	// Coupons returns all coupons in database
	Coupons(ctx context.Context) ([]*dbCoupon, error)

	Coupon(ctx context.Context, code string) (*dbCoupon, error)

	// Add inserts new coupon. If the coupon code already exists,
	// update the quantity and expiration time instead
	Add(ctx context.Context, coupon *dbCoupon) (*dbCoupon, error)

	// UpdateQuantity decreases the quantity of the specified coupon by 1.
	// This function is invoked after the coupon has been used.
	UpdateQuantity(ctx context.Context, code string) error
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
	filter := bson.M{"code": coupon.Code}
	update := bson.M{
		"$inc": bson.M{"quantity": coupon.Quantity},
		"$max": bson.M{"expiration": coupon.Expiration},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	res := coll.FindOneAndUpdate(ctx, filter, update, opts)

	var updatedCoupon *dbCoupon
	if err := res.Decode(&updatedCoupon); err != nil {
		return nil, err
	}

	return updatedCoupon, nil
}

func (r *couponStorage) UpdateQuantity(ctx context.Context, code string) error {
	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	filter := bson.M{
		"code":     code,
		"quantity": bson.M{"$gt": 0},
	}
	update := bson.D{{"$inc", bson.D{{"quantity", -1}}}}

	if err := coll.FindOneAndUpdate(ctx, filter, update).Err(); err != nil {
		return err
	}

	return nil
}

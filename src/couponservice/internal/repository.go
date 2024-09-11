package internal

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Discount is a fixed amount, not a percentage.
// CouponType_COUPON_TYPE_DISCOUNT      CouponType = 0
// CouponType_COUPON_TYPE_FREE_DELIVERY CouponType = 1
type Coupon struct {
	Types      int32
	Code       string
	Discount   int32
	Expiration int64
	Quantity   int32
}

type CouponRepository interface {

	// SaveCoupon updates the quantity and expiration time of an existing coupon.
	// If the coupon code exists, it increases the quantity and sets the expiration
	// to the latest value. If the coupon code does not exist, it inserts a new coupon.
	SaveCoupon(ctx context.Context, coupon *Coupon) error

	GetCoupon(ctx context.Context, code string) (*Coupon, error)

	ListCoupons(ctx context.Context) ([]*Coupon, error)

	// UpdateCouponQuantity decreases the quantity of the specified coupon by 1.
	UpdateCouponQuantity(ctx context.Context, code string) error
}

type couponRepository struct {
	db *mongo.Client
}

func NewCouponRepository(db *mongo.Client) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) SaveCoupon(ctx context.Context, coupon *Coupon) error {
	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	// if coupon code exists increase quantity field
	// and update with longest expiration time
	filter := bson.M{"code": coupon.Code}
	update := bson.M{
		"$inc": bson.M{"quantity": coupon.Quantity},
		"$max": bson.M{"expiration": coupon.Expiration},
	}

	if res := coll.FindOneAndUpdate(ctx, filter, update); res.Err() != nil {
		// if coupon code doesn't exists insert new coupon
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			if _, err := coll.InsertOne(ctx, coupon); err != nil {
				log.Println("Insert failed: ", err)
				return err
			}
			return nil
		}
		log.Println("Update failed: ", res.Err())
		return res.Err()
	}
	return nil
}

func (r *couponRepository) GetCoupon(ctx context.Context, code string) (*Coupon, error) {

	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")
	filter := bson.M{"code": code}

	var coupon Coupon
	if err := coll.FindOne(ctx, filter).Decode(&coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) ListCoupons(ctx context.Context) ([]*Coupon, error) {

	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")
	filter := bson.M{"quantity": bson.M{"$gt": 0}}

	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var coupons []*Coupon
	if err := cur.All(ctx, &coupons); err != nil {
		return nil, err
	}

	return coupons, nil

}

func (r *couponRepository) UpdateCouponQuantity(ctx context.Context, code string) error {
	coll := r.db.Database("coupon_database", nil).Collection("couponCollection")

	filter := bson.M{
		"code":     code,
		"quantity": bson.M{"$gt": 0},
	}

	update := bson.D{{"$inc", bson.D{{"quantity", -1}}}}
	if res := coll.FindOneAndUpdate(ctx, filter, update); res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return errors.New("coupon code does not exist or quantity is insufficient")
		}
		return res.Err()
	}
	return nil
}

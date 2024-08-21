package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderDelivery struct {
	OrderId      string     `bson:"orderId"`
	RiderId      string     `bson:"riderId"`
	IsAccepted   bool       `bson:"isAccepted"`
	AcceptedTime *time.Time `bson:"acceptedTime"`
}

type DeliveryRepo interface {
	GetOrderDeliveryById(ctx context.Context, orderId string) (*OrderDelivery, error)
	SaveOrderDelivery(ctx context.Context, orderId string) error
	UpdateOrderDelivery(ctx context.Context, orderId, riderId string, isAccepted bool) error
}

type deliveryRepo struct {
	db *mongo.Client
}

func NewDeliveryRepo(db *mongo.Client) DeliveryRepo {
	return &deliveryRepo{db: db}
}

func (r *deliveryRepo) GetOrderDeliveryById(ctx context.Context, orderId string) (*OrderDelivery, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderId": orderId}

	var result OrderDelivery

	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document found means the order is not accepted yet
			return nil, fmt.Errorf("order does not exists")
		}

		log.Println(err.Error())
		return nil, err
	}

	return &result, nil
}

func (r *deliveryRepo) SaveOrderDelivery(ctx context.Context, orderId string) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	deliveryStatus := OrderDelivery{
		OrderId:      orderId,
		RiderId:      "",
		IsAccepted:   false,
		AcceptedTime: nil,
	}

	_, err := coll.InsertOne(ctx, deliveryStatus)
	if err != nil {
		return err
	}

	return nil

}

// UpdateOrderDelivery update when Rider accepted order
func (r *deliveryRepo) UpdateOrderDelivery(ctx context.Context, orderId, riderId string, isAccepted bool) error {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderId", orderId},
	}

	update := bson.M{
		"$set": bson.M{
			"riderId":      riderId,
			"isAccepted":   isAccepted,
			"acceptedTime": time.Now(),
		},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

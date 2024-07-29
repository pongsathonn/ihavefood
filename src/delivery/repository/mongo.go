package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DeliveryRepo interface {
	GetOrderDeliveryById(ctx context.Context, orderId string) (*DeliveryStatus, error)
	SaveOrderDelivery(ctx context.Context, orderId string) error
	UpdateOrderDelivery(ctx context.Context, orderId, riderId string, isAccepted bool) error
}

type deliveryRepo struct {
	db *mongo.Client
}

// AcceptedTime is The time when the order was accepted. This field is a pointer to
// time.Time, allowing it to be nil if the order has not been accepted yet.
// Using a pointer enables us to differentiate between an unset time and a zero-value time,
// making it easier to handle cases where the acceptance time is not yet available.
type DeliveryStatus struct {
	OrderId      string     `bson:"orderId"`
	RiderId      string     `bson:"riderId"`
	IsAccepted   bool       `bson:"isAccepted"`
	AcceptedTime *time.Time `bson:"acceptedTime"`
}

func NewDeliveryRepo(db *mongo.Client) DeliveryRepo {
	return &deliveryRepo{db: db}
}

func (r *deliveryRepo) SaveOrderDelivery(ctx context.Context, orderId string) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	deliveryStatus := DeliveryStatus{
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

func (r *deliveryRepo) GetOrderDeliveryById(ctx context.Context, orderId string) (*DeliveryStatus, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderId": orderId}

	var result DeliveryStatus

	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No document found means the order is not accepted yet
			return nil, errors.New("order does not exists")
		}

		log.Printf("error finding order: %v", err)
		return nil, err
	}

	return &result, nil
}

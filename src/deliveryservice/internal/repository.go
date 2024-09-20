package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Prefix M(Model)
// prevent naming conflict with similiar struct name
//
// MOrderDelivery represent Delivery information
type MOrderDelivery struct {
	OrderId        string     `bson:"orderId"`
	RiderId        string     `bson:"riderId"`
	IsAccepted     bool       `bson:"isAccepted"`
	PickupCode     string     `bson:"pickupCode"`
	PickupLocation *MPoint    `bson:"pickupLocation"`
	Destination    *MPoint    `bson:"destination"`
	AcceptedTime   *time.Time `bson:"acceptedTime"`
}

type MPoint struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longtitude"`
}

type DeliveryRepository interface {
	GetOrderDeliveryById(ctx context.Context, orderId string) (*MOrderDelivery, error)
	SaveOrderDelivery(ctx context.Context, orderId string) error
	UpdateOrderDelivery(ctx context.Context, orderDelivery *MOrderDelivery) error
}

type deliveryRepository struct {
	db *mongo.Client
}

func NewDeliveryRepository(db *mongo.Client) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) GetOrderDeliveryById(ctx context.Context, orderId string) (*MOrderDelivery, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderId": orderId}

	var result MOrderDelivery

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

// SaveMOrderDelivery sava only orderid to database with empty other field
func (r *deliveryRepository) SaveOrderDelivery(ctx context.Context, orderId string) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	data := MOrderDelivery{
		OrderId:        orderId,
		RiderId:        "",
		PickupCode:     "",
		PickupLocation: nil,
		Destination:    nil,
		IsAccepted:     false,
		AcceptedTime:   nil,
	}

	_, err := coll.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil

}

func (r *deliveryRepository) UpdateOrderDelivery(ctx context.Context, orderDelivery *MOrderDelivery) error {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderId", orderDelivery.OrderId},
	}

	update := bson.M{
		"$set": bson.M{
			"riderId":    orderDelivery.RiderId,
			"isAccepted": orderDelivery.IsAccepted,
			"pickupCode": orderDelivery.PickupCode,
			"pickupLocation": bson.M{
				"latitude":  orderDelivery.PickupLocation.Latitude,
				"longitude": orderDelivery.PickupLocation.Longitude,
			},
			"destination": bson.M{
				"latitude":  orderDelivery.Destination.Latitude,
				"longitude": orderDelivery.Destination.Longitude,
			},
			"acceptedTime": time.Now(),
		},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

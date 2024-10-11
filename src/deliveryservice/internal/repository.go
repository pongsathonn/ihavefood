package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Prefix M(Model)
// prevent naming conflict with similiar struct name
//
// MOrderDelivery represent Delivery information
type MOrderDelivery struct {
	OrderNo        string     `bson:"orderNo"`
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
	GetOrderDelivery(ctx context.Context, orderNo string) (*MOrderDelivery, error)
	SaveOrderDelivery(ctx context.Context, orderNo string) error
	UpdateOrderDelivery(ctx context.Context, orderDelivery *MOrderDelivery) error
}

type deliveryRepository struct {
	db *mongo.Client
}

func NewDeliveryRepository(db *mongo.Client) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) GetOrderDelivery(ctx context.Context, orderNo string) (*MOrderDelivery, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderNo": orderNo}

	var result MOrderDelivery

	if err := coll.FindOne(ctx, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("order does not exists")
		}
		return nil, err
	}

	return &result, nil
}

// SaveOrderDelivery saves only the order number to the database
// and rolls back if the context is done
func (r *deliveryRepository) SaveOrderDelivery(ctx context.Context, orderNo string) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	data := MOrderDelivery{
		OrderNo:        orderNo,
		RiderId:        "",
		PickupCode:     "",
		PickupLocation: nil,
		Destination:    nil,
		IsAccepted:     false,
		AcceptedTime:   nil,
	}

	session, err := r.db.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		if _, err := coll.InsertOne(sessCtx, data); err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, nil
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *deliveryRepository) UpdateOrderDelivery(ctx context.Context, orderDelivery *MOrderDelivery) error {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderNo", orderDelivery.OrderNo},
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

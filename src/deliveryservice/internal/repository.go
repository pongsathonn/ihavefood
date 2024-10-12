package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeliveryEntity represent Delivery information
type DeliveryEntity struct {
	OrderNO        string     `bson:"orderNo"`
	RiderID        string     `bson:"riderId"`
	IsAccepted     bool       `bson:"isAccepted"`
	PickupCode     string     `bson:"pickupCode"`
	PickupLocation *Point     `bson:"pickupLocation"`
	Destination    *Point     `bson:"destination"`
	AcceptedTime   *time.Time `bson:"acceptedTime"`
}

type Point struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longtitude"`
}

type DeliveryRepository interface {
	GetDelivery(ctx context.Context, orderNO string) (*DeliveryEntity, error)
	SaveDelivery(ctx context.Context, delivery *DeliveryEntity) error
	UpdateDelivery(ctx context.Context, delivery *DeliveryEntity) error
}

type deliveryRepository struct {
	db *mongo.Client
}

func NewDeliveryRepository(db *mongo.Client) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) GetDelivery(ctx context.Context, orderNO string) (*DeliveryEntity, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderNo": orderNO}

	var result DeliveryEntity

	if err := coll.FindOne(ctx, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("order does not exists")
		}
		return nil, err
	}

	return &result, nil
}

// SaveDelivery inserts a new delivery order into the database.
// The order has not been accepted by a rider.
func (r *deliveryRepository) SaveDelivery(ctx context.Context, delivery *DeliveryEntity) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	session, err := r.db.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		_, err := coll.InsertOne(sessCtx, DeliveryEntity{
			OrderNO:    delivery.OrderNO,
			RiderID:    "",
			PickupCode: delivery.PickupCode,
			PickupLocation: &Point{
				Latitude:  delivery.PickupLocation.Latitude,
				Longitude: delivery.PickupLocation.Longitude,
			},
			Destination: &Point{
				Latitude:  delivery.Destination.Latitude,
				Longitude: delivery.Destination.Longitude,
			},
			IsAccepted:   false,
			AcceptedTime: nil,
		})
		if err != nil {
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

func (r *deliveryRepository) UpdateDelivery(ctx context.Context, delivery *DeliveryEntity) error {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderNo", delivery.OrderNO},
	}

	update := bson.M{
		"$set": bson.M{
			"riderId":      delivery.RiderID,
			"isAccepted":   delivery.IsAccepted,
			"acceptedTime": time.Now(),
		},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO doc
type DeliveryRepository interface {
	GetDelivery(ctx context.Context, orderNO string) (*DeliveryEntity, error)

	SaveDelivery(ctx context.Context, delivery *DeliveryEntity) error

	UpdateRiderAccept(ctx context.Context, delivery *DeliveryEntity) error

	UpdateRiderLocation(ctx context.Context, riderId string, point *Point) error

	IsRiderAccept(ctx context.Context, orderNO string) (bool, error)

	IsDeliver(ctx context.Context, orderNO string) (bool, error)
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
			RiderLocation: nil,
			CreatedAt:     time.Now(),
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

func (r *deliveryRepository) UpdateRiderAccept(
	ctx context.Context,
	delivery *DeliveryEntity,
) error {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderNo", delivery.OrderNO},
	}

	update := bson.D{
		"$set": bson.D{
			"riderId":               delivery.RiderID,
			"timestamps.acceptedAt": time.Now(),
		},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *deliveryRepository) UpdateRiderLocation(
	ctx context.Context,
	riderId string,
	point *Point,
) error {

	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{{"riderId", riderId}}
	update := bson.D{"$set": bson.D{"riderLocation": point}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil

}

// IsRiderAccept check whether order has accept by check delivery fields
// acceptedAt is assign or not with IsZero()
// "true" if rider has NOT accepted.
// "false" rider has already accepted the order.
func (r *deliveryRepository) IsRiderAccept(ctx context.Context, orderNO string) (bool, error) {
	coll := r.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{{"orderNo", orderNO}}

	var res DeliveryEntity
	if err := coll.FindOne(ctx, filter).Decode(&res); err != nil {
		return false, err
	}

	// AcceptedAt is zero mean rider not accept yet
	return res.AcceptedAt.IsZero(), nil
}

func (r *deliveryRepository) IsDeliver(ctx context.Context, orderNO string) (bool, error) {
	//TODO
}

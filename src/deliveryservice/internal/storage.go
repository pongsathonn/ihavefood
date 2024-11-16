package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO doc
type DeliveryStorage interface {
	Delivery(ctx context.Context, orderNO string) (*dbDelivery, error)

	// Create inserts new delivery when order is placed and return order number.
	Create(ctx context.Context, delivery *newDelivery) (string, error)

	UpdateRiderAccept(ctx context.Context, orderNO, riderID string) error

	UpdateRiderLocation(ctx context.Context, orderNO, riderID string, lo *dbPoint) error

	UpdateStatus(ctx context.Context, orderNO string, newStatus dbDeliveryStatus) error

	CheckRiderAccept(ctx context.Context, orderNO string) (bool, error)

	CheckDeliver(ctx context.Context, orderNO string) (bool, error)
}

type deliveryStorage struct {
	db *mongo.Client
}

func NewDeliveryStorage(db *mongo.Client) DeliveryStorage {
	return &deliveryStorage{db: db}
}

func (s *deliveryStorage) Delivery(ctx context.Context, orderNO string) (*dbDelivery, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderNo": orderNO}

	var delivery dbDelivery

	if err := coll.FindOne(ctx, filter).Decode(&delivery); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("order does not exists")
		}
		return nil, err
	}

	return &delivery, nil
}

// Create inserts a new delivery order into the database. The order has not been accepted by a rider.
func (s *deliveryStorage) Create(ctx context.Context, delivery *newDelivery) (string, error) {

	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	res, err := coll.InsertOne(ctx, delivery)
	if err != nil {
		return "", err
	}
	orderNO := res.InsertedID.(string)

	return orderNO, nil
}

func (s *deliveryStorage) UpdateRiderAccept(ctx context.Context, orderNO, riderID string) error {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{{"orderNo", orderNO}}

	update := bson.D{{"$set", bson.D{
		{"riderId", riderID},
		{"status", ACCEPTED},
		{"timestamps.acceptTime", time.Now()},
	}}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *deliveryStorage) UpdateRiderLocation(ctx context.Context, orderNO, riderID string, lo *dbPoint) error {

	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{
		{"orderNo", orderNO},
		{"riderId", riderID},
	}

	update := bson.D{{"$set", bson.D{
		{"riderLocation.latitude", lo.Latitude},
		{"riderLocation.longitude", lo.Longitude},
	}}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil

}

func (s *deliveryStorage) UpdateStatus(ctx context.Context, orderNO string, newStatus dbDeliveryStatus) error {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{{"orderNo", orderNO}}
	update := bson.D{{"$set", bson.D{{"status", newStatus}}}}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *deliveryStorage) CheckRiderAccept(ctx context.Context, orderNO string) (bool, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.D{{"orderNo", orderNO}}

	var res dbDelivery
	if err := coll.FindOne(ctx, filter).Decode(&res); err != nil {
		return false, err
	}

	if res.Status == ACCEPTED {
		return false, nil
	}

	return true, nil
}

func (s *deliveryStorage) CheckDeliver(ctx context.Context, orderNO string) (bool, error) {
	//TODO implement
	return false, nil
}

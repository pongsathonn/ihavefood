package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO doc
type DeliveryStorage interface {
	Delivery(ctx context.Context, orderID string) (*dbDelivery, error)

	// Create inserts new delivery when order is placed and return order number.
	Create(ctx context.Context, delivery *newDelivery) (string, error)

	UpdateRiderAccept(ctx context.Context, orderID, riderID string) (string, error)

	UpdateRiderLocation(ctx context.Context, orderID, riderID string, lo *dbPoint) (string, error)

	UpdateStatus(ctx context.Context, orderID string, newStatus dbDeliveryStatus) (string, error)

	CheckRiderAccept(ctx context.Context, orderID string) (bool, error)

	CheckDeliver(ctx context.Context, orderID string) (bool, error)
}

type deliveryStorage struct {
	db *mongo.Client
}

func NewDeliveryStorage(db *mongo.Client) DeliveryStorage {
	return &deliveryStorage{db: db}
}

func (s *deliveryStorage) Delivery(ctx context.Context, orderID string) (*dbDelivery, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	filter := bson.M{"orderId": orderID}

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
	orderID := res.InsertedID.(string)

	return orderID, nil
}

func (s *deliveryStorage) UpdateRiderAccept(ctx context.Context, orderID, riderID string) (string, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return "", err
	}

	update := bson.D{{"$set", bson.D{
		{"riderId", riderID},
		{"status", ACCEPTED},
		{"timestamps.acceptTime", time.Now()},
	}}}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.MatchedCount == 0 {
		return "", errors.New("orer not found")
	}

	updatedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("convert ID to primitive.ObjectID failed")
	}

	return updatedID.Hex(), nil
}

func (s *deliveryStorage) UpdateRiderLocation(ctx context.Context, orderID, riderID string, lo *dbPoint) (string, error) {

	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return "", err
	}

	filter := bson.D{
		{"_id", ID},
		{"riderId", riderID},
	}

	update := bson.D{{"$set", bson.D{
		{"riderLocation.latitude", lo.Latitude},
		{"riderLocation.longitude", lo.Longitude},
	}}}

	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", err
	}

	if res.MatchedCount == 0 {
		return "", errors.New("orderID or riderID not match")
	}

	updatedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return updatedID.Hex(), nil

}

func (s *deliveryStorage) UpdateStatus(ctx context.Context, orderID string, newStatus dbDeliveryStatus) (string, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return "", err
	}

	update := bson.D{{"$set", bson.D{{"status", newStatus}}}}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.MatchedCount == 0 {
		return "", errors.New("order not found")
	}

	updatedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return updatedID.Hex(), nil
}

func (s *deliveryStorage) CheckRiderAccept(ctx context.Context, orderID string) (bool, error) {
	coll := s.db.Database("delivery_database", nil).Collection("deliveryCollection")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return false, err
	}

	filter := bson.D{{"_id", ID}}

	var res dbDelivery
	if err := coll.FindOne(ctx, filter).Decode(&res); err != nil {
		return false, err
	}

	if res.Status == ACCEPTED {
		return false, nil
	}

	return true, nil
}

func (s *deliveryStorage) CheckDeliver(ctx context.Context, orderID string) (bool, error) {
	//TODO implement
	return false, nil
}

package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderStorage interface {

	// Retrieves the order history for a specified customer by username
	ListPlaceOrders(ctx context.Context, customerID string) ([]*dbPlaceOrder, error)

	GetPlaceOrder(ctx context.Context, orderID string) (*dbPlaceOrder, error)

	// Create inserts new place order into database and returns the order number.
	Create(ctx context.Context, newOrder *newPlaceOrder) (string, error)

	UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (bool, error)

	UpdatePaymentStatus(ctx context.Context, orderID string, status dbPaymentStatus) (bool, error)

	DeletePlaceOrder(ctx context.Context, orderID string) error
}

type orderStorage struct {
	client *mongo.Client
}

func NewOrderStorage(client *mongo.Client) OrderStorage {
	return &orderStorage{client: client}
}

func (s *orderStorage) Create(ctx context.Context, newOrder *newPlaceOrder) (string, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	res, err := coll.InsertOne(ctx, newOrder)
	if err != nil {
		return "", err
	}

	orderID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to convert insertedID to primitive.objectID")
	}

	return orderID.Hex(), nil

}

func (s *orderStorage) ListPlaceOrders(ctx context.Context, customerID string) ([]*dbPlaceOrder, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	cur, err := coll.Find(ctx, bson.M{"customerID": customerID})
	if err != nil {
		return nil, err
	}

	var orders []*dbPlaceOrder

	for cur.Next(ctx) {
		var order dbPlaceOrder
		if err := cur.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *orderStorage) GetPlaceOrder(ctx context.Context, orderID string) (*dbPlaceOrder, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}

	var order dbPlaceOrder
	if err := coll.FindOne(ctx, bson.D{{Key: "_id", Value: ID}}).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *orderStorage) UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (bool, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return false, err
	}

	var timestamp string
	timestamp = "timestamps.updateTime"

	// Updating to "PENDING" will result in an error, as it is the default status.
	if status == OrderStatus_PENDING {
		return false, errors.New("pending is default status can not be updated")
	}

	if status == OrderStatus_DELIVERED {
		timestamp = "timestamps.completeTime"
	}

	now := time.Now()
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "orderStatus", Value: status},
			{Key: timestamp, Value: now},
		}},
	}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return false, err
	}

	if res.MatchedCount == 0 {
		return false, errors.New("order not found")
	}

	return true, nil
}

func (s *orderStorage) UpdatePaymentStatus(ctx context.Context, orderID string, status dbPaymentStatus) (bool, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return false, err
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "paymentStatus", Value: status},
		}},
	}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return false, err
	}

	if res.MatchedCount == 0 {
		return false, errors.New("order not found")
	}
	return true, nil
}

func (s *orderStorage) DeletePlaceOrder(ctx context.Context, orderID string) error {
	coll := s.client.Database("db").Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	_, err = coll.DeleteOne(ctx, bson.M{"_id": ID})
	if err != nil {
		return err
	}

	return nil
}

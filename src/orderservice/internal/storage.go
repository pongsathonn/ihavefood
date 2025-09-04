package internal

import (
	"context"
	"errors"
	"log/slog"
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

	UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (string, error)

	UpdatePaymentStatus(ctx context.Context, orderID string, status dbPaymentStatus) (string, error)

	//DeletePlaceOrder(ctx context.Context, orderID string) error
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

	cur, err := coll.Find(ctx, bson.D{{"customerID", customerID}})
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
	if err := coll.FindOne(ctx, bson.D{{"_id", ID}}).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *orderStorage) UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (string, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return "", err
	}

	var timestamp string
	timestamp = "timestamps.updateTime"

	// Updating to "PENDING" will result in an error, as it is the default status.
	if status == OrderStatus_PENDING {
		return "", errors.New("pending is default status can not be updated")
	}

	if status == OrderStatus_DELIVERED {
		timestamp = "timestamps.completeTime"
	}

	now := time.Now()

	update := bson.D{
		{"$set", bson.D{
			{"orderStatus", status},
			{timestamp, now},
		}},
	}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		slog.Info("order number %s not found", orderID)
		return "", errors.New("order not found")
	}

	updatedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to convert upsertedID to primitive.ObjectID")
	}

	return updatedID.Hex(), nil
}

func (s *orderStorage) UpdatePaymentStatus(ctx context.Context, orderID string, status dbPaymentStatus) (string, error) {

	coll := s.client.Database("db", nil).Collection("orders")

	ID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return "", err
	}

	if status == PaymentStatus_UNPAID {
		return "", errors.New("unpaid is default status")
	}

	update := bson.D{{"$set", bson.D{
		{"paymentStatus", status},
	}}}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		slog.Info("order number %s not found", orderID)
		return "", errors.New("order not found")
	}

	updatedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to convert upsertedID to primitive.ObjectID")
	}

	return updatedID.Hex(), nil
}

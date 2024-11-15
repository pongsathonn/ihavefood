package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderStorage interface {
	SaveNewPlaceOrder(ctx context.Context, in *dbPlaceOrder) (*dbPlaceOrder, error)

	// Retrieves the order history for a specified user by username
	PlaceOrders(ctx context.Context, username string) ([]*dbPlaceOrder, error)

	PlaceOrder(ctx context.Context, orderNO string) (*dbPlaceOrder, error)

	UpdateOrderStatus(ctx context.Context, orderNO string, status dbOrderStatus) (*dbPlaceOrder, error)

	UpdatePaymentStatus(ctx context.Context, orderNO string, status dbPaymentStatus) (*dbPlaceOrder, error)

	//DeletePlaceOrder(ctx context.Context, orderNO string) error
}

type orderStorage struct {
	client *mongo.Client
}

func NewOrderStorage(client *mongo.Client) OrderStorage {
	return &orderStorage{client: client}
}

func (s *orderStorage) SaveNewPlaceOrder(ctx context.Context, in *dbPlaceOrder) (*dbPlaceOrder, error) {

	coll := s.client.Database("order_database", nil).Collection("orderCollection")

	res, err := coll.InsertOne(ctx, in)
	if err != nil {
		return nil, err
	}

	var order dbPlaceOrder
	if err := coll.FindOne(ctx, bson.D{{"_id", res.InsertedID}}).Decode(order); err != nil {
		return nil, err
	}

	return &order, nil

}

func (s *orderStorage) PlaceOrders(ctx context.Context, username string) ([]*dbPlaceOrder, error) {

	coll := s.client.Database("order_database", nil).Collection("orderCollection")

	cur, err := coll.Find(ctx, bson.D{{"username", username}})
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

func (s *orderStorage) PlaceOrder(ctx context.Context, orderNO string) (*dbPlaceOrder, error) {
	coll := s.client.Database("order_database", nil).Collection("orderCollection")

	ID, err := primitive.ObjectIDFromHex(orderNO)
	if err != nil {
		return nil, err
	}

	var order dbPlaceOrder
	if err := coll.FindOne(ctx, bson.D{{"_id", ID}}).Decode(order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *orderStorage) UpdateOrderStatus(ctx context.Context, orderNo string,
	status dbOrderStatus) (*dbPlaceOrder, error) {

	coll := s.client.Database("order_database", nil).Collection("orderCollection")

	orderNumber, err := primitive.ObjectIDFromHex(orderNo)
	if err != nil {
		return nil, err
	}

	var timestamp string
	timestamp = "timestamps.updateTime"

	// Updating to "PENDING" will result in an error, as it is the default status.
	if status == OrderStatus_PENDING {
		return nil, errors.New("pending is default status can not be updated")
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

	filter := bson.D{{"_id", orderNumber}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var order dbPlaceOrder
	if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *orderStorage) UpdatePaymentStatus(ctx context.Context, orderNo string,
	status dbPaymentStatus) (*dbPlaceOrder, error) {

	coll := s.client.Database("order_database", nil).Collection("orderCollection")

	orderNumber, err := primitive.ObjectIDFromHex(orderNo)
	if err != nil {
		return nil, err
	}

	if status == PaymentStatus_UNPAID {
		return nil, errors.New("unpaid is default status")
	}

	filter := bson.D{{"_id", orderNumber}}

	update := bson.D{{"$set", bson.D{
		{"paymentStatus", status},
	}}}

	var order dbPlaceOrder
	if err := coll.FindOneAndUpdate(ctx, filter, update).Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}

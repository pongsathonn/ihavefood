package data

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

type OrderDatabase interface {
	SavePlaceOrder() error
}

func NewOrderDatabase(conn *mongo.Client) OrderDatabase {
	return &orderDatabase{conn: conn}
}

type orderDatabase struct {
	conn *mongo.Client
}

func (od *orderDatabase) SavePlaceOrder() error {
	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	res, err := coll.InsertOne(context.Background(), nil)
	if err != nil {
		return err
	}

	log.Println(res.InsertedID)
	return nil
}

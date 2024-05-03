package data

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/food-delivery/src/order/genproto"
)

type OrderDatabase interface {
	SavePlaceOrder(*pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error)
	FindPlaceOrder(username string) (string, error)
}

func NewOrderDatabase(conn *mongo.Client) OrderDatabase {
	return &orderDatabase{conn: conn}
}

type orderDatabase struct {
	conn *mongo.Client
}

type PlaceOrder struct {
	OrderId         primitive.ObjectID  `bson:"_id"`
	TrackingId      primitive.ObjectID  `bson:"tracking_id"`
	Username        string              `bson:"username"`
	Email           string              `bson:"email"`
	OrderCost       int32               `bson:"order_cost"`
	Menus           []*pb.Menu          `bson:"menus"`
	DeliveryAddress *pb.DeliveryAddress `bson:"delivery_address"`
}

func (od *orderDatabase) FindPlaceOrder(username string) (string, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	// result will be this type
	var res bson.M

	// use this find data in mongo
	x := bson.D{{"username", username}}
	err := coll.FindOne(context.TODO(), x).Decode(&res)
	if err == mongo.ErrNoDocuments {
		return "", fmt.Errorf("documents not found")
	}
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		return "", err
	}

	resp := fmt.Sprintf("%s\n", jsonData)

	return resp, nil

}

func (od *orderDatabase) SavePlaceOrder(po *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	placeOrder := PlaceOrder{
		OrderId:         primitive.NewObjectID(),
		TrackingId:      primitive.NewObjectID(),
		Username:        po.Username,
		Email:           po.Email,
		OrderCost:       po.OrderCost,
		Menus:           po.Menus,
		DeliveryAddress: po.Address,
	}

	_, err := coll.InsertOne(context.TODO(), &placeOrder)
	if err != nil {
		return nil, err
	}

	return &pb.PlaceOrderResponse{
		OrderId:         placeOrder.OrderId.Hex(),
		OrderTrackingId: placeOrder.TrackingId.Hex(),
	}, nil

}

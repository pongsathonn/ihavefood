package data

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/food-delivery/src/order/genproto"
)

type OrderRepo interface {
	ListPlaceOrder(username string) (*pb.ListUserPlaceOrderResponse, error)
	SavePlaceOrder(*pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error)
}

func NewOrderRepo(conn *mongo.Client) OrderRepo {
	return &orderRepo{conn: conn}
}

type orderRepo struct {
	conn *mongo.Client
}

// this PlaceOrder use as model for "INSERT" to mongodb
// to make this model match with protobuff json name should be same as bson name tag
// if not omitempty at id Mongo will use zero as id ( when insert )
type PlaceOrder struct {
	OrderId         primitive.ObjectID  `bson:"_id,omitempty"`
	TrackingId      primitive.ObjectID  `bson:"orderTrackingId" `
	Username        string              `bson:"username"`
	Total           int32               `bson:"orderCost"`
	CouponCode      string              `bson:"total"`
	Menus           []*pb.Menu          `bson:"menus"`
	DeliveryAddress *pb.DeliveryAddress `bson:"address"`
	ContactInfo     *pb.ContactInfo     `bson:"contactInfo"`
	PaymentMethod   pb.PaymentMethod    `bson:"paymentMethod"`
	PaymentStatus   pb.PaymentStatus    `bson:"paymentStatus"`
	OrderStatus     pb.OrderStatus      `bson:"orderStatus"`
}

// list all user's placeorder by username ( like query placeorder history )
func (od *orderRepo) ListPlaceOrder(username string) (*pb.ListUserPlaceOrderResponse, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	filter := bson.D{{"username", username}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var orders []*pb.Order

	if err := cursor.All(context.TODO(), &orders); err != nil {
		log.Fatal(err)
	}

	log.Println(orders)

	//FIXME  orderId (json ) shouldn't be empty. It shoule map with mongo _id
	return &pb.ListUserPlaceOrderResponse{
		Orders: orders,
	}, nil
}

func (od *orderRepo) SavePlaceOrder(in *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	// still need to send avaliable true
	placeOrder := PlaceOrder{
		OrderId:         primitive.NewObjectID(),
		TrackingId:      primitive.NewObjectID(),
		Username:        in.Username,
		Total:           in.Total,
		Menus:           in.Menus,
		CouponCode:      in.CouponCode,
		DeliveryAddress: in.Address,
		ContactInfo:     in.Contact,
		PaymentMethod:   in.PaymentMethod,
		PaymentStatus:   pb.PaymentStatus_UNPAID,
		OrderStatus:     pb.OrderStatus_PENDING,
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

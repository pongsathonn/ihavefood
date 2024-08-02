package data

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

type OrderRepo interface {
	PlaceOrder(username string) (*pb.ListUserPlaceOrderResponse, error)
	SavePlaceOrder(*pb.PlaceOrderRequest) (*SavePlaceOrderResponse, error)
}

func NewOrderRepo(conn *mongo.Client) OrderRepo {
	return &orderRepo{conn: conn}
}

// this PlaceOrder use as model for "INSERT" and "QUERY" with mongodb
// to make this model match with protobuff json name should be same as bson name tag
// if not omitempty at id Mongo will use zero as id ( when insert )
type PlaceOrder struct {
	OrderId        primitive.ObjectID `bson:"_id,omitempty"`
	TrackingId     primitive.ObjectID `bson:"orderTrackingId" `
	Username       string             `bson:"username"`
	RestaurantName string             `bson:"restaurantName"`
	Menus          []*pb.Menu         `bson:"menus"`
	CouponCode     string             `bson:"couponCode"`
	CouponDiscount int32              `bson:"couponDiscount"`
	DeliveryFee    int32              `bson:"deliveryFee"`
	Total          int32              `bson:"total"`
	UserAddress    *pb.Address        `bson:"userAddress"`
	ContactInfo    *pb.ContactInfo    `bson:"contactInfo"`
	PaymentMethod  pb.PaymentMethod   `bson:"paymentMethod"`
	PaymentStatus  pb.PaymentStatus   `bson:"paymentStatus"`
	OrderStatus    pb.OrderStatus     `bson:"orderStatus"`
}

type SavePlaceOrderResponse struct {
	OrderId         string
	OrderTrackingId string
	PaymentStatus   pb.PaymentStatus
	OrderStatus     pb.OrderStatus
}

type orderRepo struct {
	conn *mongo.Client
}

// list all user's placeorder by username ( like query placeorder history )
func (od *orderRepo) PlaceOrder(username string) (*pb.ListUserPlaceOrderResponse, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	filter := bson.D{{"username", username}}
	cur, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var placeOrders []*pb.PlaceOrder

	for cur.Next(context.TODO()) {
		var po PlaceOrder
		if err := cur.Decode(&po); err != nil {
			return nil, err
		}

		placeOrder := &pb.PlaceOrder{
			OrderId:         po.OrderId.Hex(),
			OrderTrackingId: po.TrackingId.Hex(),
			Username:        po.Username,
			RestaurantName:  po.RestaurantName,
			Menus:           po.Menus,
			CouponCode:      po.CouponCode,
			CouponDiscount:  po.CouponDiscount,
			DeliveryFee:     po.DeliveryFee,
			Total:           po.Total,
			Address:         po.UserAddress,
			Contact:         po.ContactInfo,
			PaymentMethod:   po.PaymentMethod,
			PaymentStatus:   po.PaymentStatus,
			OrderStatus:     po.OrderStatus,
		}
		placeOrders = append(placeOrders, placeOrder)
	}

	return &pb.ListUserPlaceOrderResponse{PlaceOrders: placeOrders}, nil
}

// TODO generate tracking id
func (od *orderRepo) SavePlaceOrder(in *pb.PlaceOrderRequest) (*SavePlaceOrderResponse, error) {

	coll := od.conn.Database("order_database", nil).Collection("orderCollection")

	po := PlaceOrder{
		TrackingId:     primitive.NewObjectID(),
		RestaurantName: in.RestaurantName,
		Username:       in.Username,
		CouponCode:     in.CouponCode,
		CouponDiscount: in.CouponDiscount,
		Menus:          in.Menus,
		DeliveryFee:    in.DeliveryFee,
		Total:          in.Total,
		UserAddress:    in.Address,
		ContactInfo:    in.Contact,
		PaymentMethod:  in.PaymentMethod,
		PaymentStatus:  pb.PaymentStatus_UNPAID,
		OrderStatus:    pb.OrderStatus_PENDING,
	}

	res, err := coll.InsertOne(context.TODO(), &po)
	if err != nil {
		return nil, err
	}

	orderId, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("error assert type")
	}

	return &SavePlaceOrderResponse{
		OrderId:         orderId.Hex(),
		OrderTrackingId: po.TrackingId.Hex(),
		PaymentStatus:   pb.PaymentStatus_UNPAID,
		OrderStatus:     pb.OrderStatus_PENDING,
	}, nil

}

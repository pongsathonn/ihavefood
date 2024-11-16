package internal

import (
	"time"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
)

type newDelivery struct {

	// OrderNO must be insert as _id
	OrderNO        string           `bson:"_id"`
	PickupCode     string           `bson:"pickupCode"`
	PickupLocation *dbPoint         `bson:"pickupLocation"`
	Destination    *dbPoint         `bson:"destination"`
	Status         dbDeliveryStatus `bson:"status"`
	Timestamp      *dbTimeStamp     `bson:"timestamp"`
}

// dbDelivery represent delivery information for an order
type dbDelivery struct {

	// OrderNO must be insert as _id
	OrderNO string `bson:"_id"`

	// PickupCode is code 3 digit for rider pickup
	PickupCode string `bson:"pickupCode"`

	// PickupLocation is Restaurant address
	PickupLocation *dbPoint `bson:"pickupLocation"`

	// Destination is User address
	Destination *dbPoint `bson:"destination"`

	// RiderID who accept the order
	RiderID string `bson:"riderId"`

	// Current rider location
	RiderLocation *dbPoint `bson:"riderLocation"`

	// Delivery status
	Status dbDeliveryStatus `bson:"status"`

	Timestamp *dbTimeStamp `bson:"timestamp"`
}

type dbPoint struct {
	Latitude float64 `bson:"latitude"`

	Longitude float64 `bson:"longitude"`
}

type dbDeliveryStatus int32

const (
	// UNACCEPTED indicates the rider has not yet accepted the order.
	UNACCEPT dbDeliveryStatus = 0

	// ACCEPTED indicates the rider has accepted the order.
	ACCEPTED dbDeliveryStatus = 1

	// DELIVERED indicates the order has been delivered by the rider.
	DELIVERED dbDeliveryStatus = 2
)

type dbTimeStamp struct {
	// CreateTime is the timestamp when the DeliveryService receives
	// a new order.
	CreateTime time.Time `bson:"createTime"`

	// AcceptTime is the timestamp when the rider accepts the order.
	AcceptTime time.Time `bson:"acceptTime"`

	// DeliverTime is the timestamp when the order is delivered.
	DeliverTime time.Time `bson:"deliverTime"`
}

//------------- EXAMPLE DATA ---------------------------------

// [ Chaing Mai district ]
var example = map[string]*pb.Point{
	"Mueang":    &pb.Point{Latitude: 18.7883, Longitude: 98.9853},
	"Hang Dong": &pb.Point{Latitude: 18.6870, Longitude: 98.8897},
	"San Sai":   &pb.Point{Latitude: 18.8578, Longitude: 99.0631},
	"Mae Rim":   &pb.Point{Latitude: 18.8998, Longitude: 98.9311},
	"Doi Saket": &pb.Point{Latitude: 18.8482, Longitude: 99.1403},
}

var riders = []*pb.Rider{
	{Id: "001", Name: "Messi", PhoneNumber: "0846851976"},
	{Id: "002", Name: "Ronaldo", PhoneNumber: "0987858487"},
	{Id: "003", Name: "Neymar", PhoneNumber: "0684321352"},
	{Id: "004", Name: "pogba", PhoneNumber: "0868549858"},
	{Id: "005", Name: "Halaand", PhoneNumber: "0932515487"},
}

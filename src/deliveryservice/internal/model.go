package internal

import (
	"time"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
)

type DeliveryEntity struct {
	OrderNO string `bson:"orderNo"`

	// PickupCode is code 3 digit for rider pickup
	PickupCode string `bson:"pickupCode"`

	// PickupLocation is Restaurant address
	PickupLocation *Point `bson:"pickupLocation"`

	// Destination is User address
	Destination *Point `bson:"destination"`

	// RiderID who accept the order
	RiderID string `bson:"riderId"`

	// RiderLocation current rider location
	RiderLocation *Point `bson:"riderLocation"`

	// CreatedAt is the timestamp when insert new order.
	CreatedAt time.Time `bson:"assignedAt"`

	// AcceptedAt is the timestamp when the rider accepts the order.
	AcceptedAt time.Time `bson:"acceptedAt"`

	// DeliveredAt is the timestamp when the rider delivers the order.
	DeliveredAt time.Time `bson:"deliveredAt"`
}

type Point struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longtitude"`
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

package internal

import pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"

// Example data ( Chaing Mai district )
var example = map[string]*pb.Point{
	"Mueang":    &pb.Point{Latitude: 18.7883, Longitude: 98.9853},
	"Hang Dong": &pb.Point{Latitude: 18.6870, Longitude: 98.8897},
	"San Sai":   &pb.Point{Latitude: 18.8578, Longitude: 99.0631},
	"Mae Rim":   &pb.Point{Latitude: 18.8998, Longitude: 98.9311},
	"Doi Saket": &pb.Point{Latitude: 18.8482, Longitude: 99.1403},
}

// Example data for riders
var riders = []*pb.Rider{
	{Id: "001", Name: "Messi", PhoneNumber: "0846851976"},
	{Id: "002", Name: "Ronaldo", PhoneNumber: "0987858487"},
	{Id: "003", Name: "Neymar", PhoneNumber: "0684321352"},
	{Id: "004", Name: "pogba", PhoneNumber: "0868549858"},
	{Id: "005", Name: "Halaand", PhoneNumber: "0932515487"},
}

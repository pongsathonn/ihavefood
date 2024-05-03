package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	database "github.com/pongsathonn/food-delivery/src/order/data"
	pb "github.com/pongsathonn/food-delivery/src/order/genproto"
)

// TODO: use environment variable instead
var (
	mongoPort = 27017
	uri       = fmt.Sprintf("mongodb://kenmilez:mypassword@localhost:%d", mongoPort)

	serverPort    = 4010
	serverAddress = fmt.Sprintf("localhost:%d", serverPort)
)

type orderServer struct {
	pb.UnimplementedOrderServiceServer
	db database.OrderDatabase
}

func NewOrderServer(db database.OrderDatabase) *orderServer {
	return &orderServer{db: db}
}

func (or *orderServer) PlaceOrder(ctx context.Context, in *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	res, err := or.db.SavePlaceOrder(in)
	if err != nil {
		log.Println(err)
	}

	/*
		TODO
		- publish CreatePlaceOrderEvent
	*/

	return res, nil
}

func initMongoConnection() *mongo.Client {

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Ping: ", err)
	}

	return client
}

func main() {

	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Println("listen on addr :", err)
	}

	// wire up
	orderDB := database.NewOrderDatabase(initMongoConnection())
	orSrv := NewOrderServer(orderDB)

	var srv *grpc.Server
	srv = grpc.NewServer()

	pb.RegisterOrderServiceServer(srv, orSrv)
	log.Println("server is running on port :", serverPort)

	if err := srv.Serve(lis); err != nil {
		log.Println("err lis :", err)
	}
}

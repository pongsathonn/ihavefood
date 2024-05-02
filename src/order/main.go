package main

import (
	"context"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	database "github.com/pongsathonn/food-delivery/src/order/data"
	pb "github.com/pongsathonn/food-delivery/src/order/genproto"
)

// TODO: use environment variable instead
const urlx = "mongodb://kenmilez:mypassword@localhost:27017"

type orderServer struct {
	pb.UnimplementedOrderServiceServer
	db database.OrderDatabase
}

func NewOrderServer(db database.OrderDatabase) *orderServer {
	return &orderServer{db: db}
}

type PlaceOrder struct {
	Username  string
	Email     string
	OrderCost int32
	Menus     []*pb.Menu
	Address   *pb.DeliveryAddress
}

/*
TODO : write following this logic
	1) receive PlaceOrder request with payload
	2) save order to order_database
	3) response order_id from database
	4) create PlaceOrderEvent to event bus
*/

func (or *orderServer) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	_ = &PlaceOrder{
		Username:  req.Username,
		OrderCost: req.OrderCost,
		Menus:     req.Menus,
		Email:     req.Email,
		Addr:      req.Address,
	}

	err := or.db.SavePlaceOrder()
	if err != nil {
		log.Println(err)
	}

	return nil, nil
}

func initMongoConnection() *mongo.Client {

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(urlx))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Ping: ", err)
	}

	return client
}

func main() {
	var srv *grpc.Server
	srv = grpc.NewServer()

	lis, err := net.Listen("tcp", "localhost:4010")
	if err != nil {
		log.Fatal(err)
	}

	// wire up
	orderDB := database.NewOrderDatabase(initMongoConnection())
	NewOrderServer(orderDB)

	pb.RegisterOrderServiceServer(srv, nil)
	log.Println("server is running")

	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

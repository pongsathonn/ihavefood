package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/pongsathonn/ihavefood/src/orderservice/internal"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	repository := internal.NewOrderRepository(initMongoClient())
	rabbitmq := internal.NewRabbitMQ(initRabbitMQ())
	orderService := internal.NewOrderService(repository, rabbitmq)
	startGRPCServer(orderService)
}

func initRabbitMQ() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("ORDER_AMQP_USER"),
		os.Getenv("ORDER_AMQP_PASS"),
		os.Getenv("ORDER_AMQP_HOST"),
		os.Getenv("ORDER_AMQP_PORT"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func initMongoClient() *mongo.Client {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/order_database?authSource=admin",
		os.Getenv("ORDER_MONGO_USER"),
		os.Getenv("ORDER_MONGO_PASS"),
		os.Getenv("ORDER_MONGO_HOST"),
		os.Getenv("ORDER_MONGO_PORT"),
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Database("order_database").CreateCollection(context.TODO(), "orderCollection")
	if err != nil {
		var alreayExistsColl mongo.CommandError
		if !errors.As(err, &alreayExistsColl) {
			log.Fatal(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	return client

}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(s *internal.OrderService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", os.Getenv("ORDER_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, s)

	log.Printf("order service is running on port %s\n", os.Getenv("ORDER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

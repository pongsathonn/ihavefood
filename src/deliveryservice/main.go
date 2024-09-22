package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
	"github.com/pongsathonn/ihavefood/src/deliveryservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	conn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	client, err := initMongoDB()
	if err != nil {
		log.Fatal(err)
	}

	rabbitmq := internal.NewRabbitMQ(conn)
	repository := internal.NewDeliveryRepository(client)

	deliveryService := internal.NewDeliveryService(
		rabbitmq,
		repository,
	)

	// Start the order assignment process in a separate goroutine
	go deliveryService.DeliveryAssignment()

	startGRPCServer(deliveryService)
}

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("DELIVERY_AMQP_USER"),
		os.Getenv("DELIVERY_AMQP_PASS"),
		os.Getenv("DELIVERY_AMQP_HOST"),
		os.Getenv("DELIVERY_AMQP_PORT"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// initRepository initializes the MongoDB connection and returns the delivery repository instance
func initMongoDB() (*mongo.Client, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/delivery_database?authSource=admin",
		os.Getenv("DELIVERY_MONGO_USER"),
		os.Getenv("DELIVERY_MONGO_PASS"),
		os.Getenv("DELIVERY_MONGO_HOST"),
		os.Getenv("DELIVERY_MONGO_PORT"),
	)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	return client, nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(s *internal.DeliveryService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", os.Getenv("DELIVERY_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterDeliveryServiceServer(grpcServer, s)

	log.Printf("delivery service is running on port %s\n", os.Getenv("DELIVERY_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

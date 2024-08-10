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

	"github.com/pongsathonn/ihavefood/src/deliveryservice/rabbitmq"
	"github.com/pongsathonn/ihavefood/src/deliveryservice/repository"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("DELIVERY_AMQP_USER", "donkadmin"),
		getEnv("DELIVERY_AMQP_PASS", "donkpassword"),
		getEnv("DELIVERY_AMQP_HOST", "localhost"),
		getEnv("DELIVERY_AMQP_PORT", "5672"),
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
		getEnv("DELIVERY_MONGO_USER", "donkadmin"),
		getEnv("DELIVERY_MONGO_PASS", "donkpassword"),
		getEnv("DELIVERY_MONGO_HOST", "localhost"),
		getEnv("DELIVERY_MONGO_PORT", "27017"),
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
func startGRPCServer(s *deliveryService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("DELIVERY_SERVER_PORT", "5555"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterDeliveryServiceServer(grpcServer, s)

	log.Printf("delivery service is running on port %s\n", getEnv("DELIVERY_SERVER_PORT", "5555"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// getEnv fetches an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("value of %s not set\n", key)
	return defaultValue
}

func main() {

	conn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}
	rb := rabbitmq.NewRabbitMQ(conn)

	client, err := initMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	rp := repository.NewDeliveryRepo(client)

	d := NewDeliveryService(rb, rp)

	// Start the order assignment process in a separate goroutine
	go d.orderAssignment()

	startGRPCServer(d)
}

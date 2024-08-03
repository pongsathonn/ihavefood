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

	"github.com/pongsathonn/ihavefood/src/deliveryservice/pubsub"
	"github.com/pongsathonn/ihavefood/src/deliveryservice/repository"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// initPubSub initializes the RabbitMQ connection and returns the pubsub instance
func initPubSub() (pubsub.RabbitMQ, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("DELIVERY_AMQP_USER", "donkadmin"),
		getEnv("DELIVERY_AMQP_PASS", "donkpassword"),
		getEnv("DELIVERY_AMQP_HOST", "localhost"),
		getEnv("DELIVERY_AMQP_PORT", "5672"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return pubsub.NewRabbitMQ(conn), nil
}

// initRepository initializes the MongoDB connection and returns the delivery repository instance
func initRepository(ctx context.Context) (repository.DeliveryRepo, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/delivery_database?authSource=admin",
		getEnv("DELIVERY_MONGO_USER", "donkadmin"),
		getEnv("DELIVERY_MONGO_PASS", "donkpassword"),
		getEnv("DELIVERY_MONGO_HOST", "localhost"),
		getEnv("DELIVERY_MONGO_PORT", "27017"),
	)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return repository.NewDeliveryRepo(client), nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(ds *delivery) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("DELIVERY_SERVER_PORT", "5555"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	s := grpc.NewServer()
	pb.RegisterDeliveryServiceServer(s, ds)

	log.Printf("Delivery service is running on port %s\n", getEnv("DELIVERY_SERVER_PORT", "5555"))

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// getEnv fetches an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependencies
	ps, err := initPubSub()
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}

	rp, err := initRepository(ctx)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	d := newDelivery(ps, rp)

	// Start the order assignment process in a separate goroutine
	go ds.orderAssignment()

	// Set up and start the gRPC server
	startGRPCServer(d)
}

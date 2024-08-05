package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/pongsathonn/ihavefood/src/orderservice/rabbitmq"
	"github.com/pongsathonn/ihavefood/src/orderservice/repository"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("ORDER_AMQP_USER"),
		os.Getenv("ORDER_AMQP_PASS"),
		os.Getenv("ORDER_AMQP_HOST"),
		os.Getenv("ORDER_AMQP_PORT"),
	)

	//uri := "amqp://donkadmin:donkpassword@rabbitmqx:5672"
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func initMongoClient() (*mongo.Client, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/order_database?authSource=admin",
		os.Getenv("ORDER_MONGO_USER"),
		os.Getenv("ORDER_MONGO_PASS"),
		os.Getenv("ORDER_MONGO_HOST"),
		os.Getenv("ORDER_MONGO_PORT"),
	)

	//uri := "mongodb://donkadmin:donkpassword@orderdb:27017/order_database?authSource=admin"

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Create collection if not exists
	err = client.Database("order_database").CreateCollection(context.TODO(), "orderCollection")
	if err != nil {
		//TODO if exists pass
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(s *order) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("ORDER_SERVER_PORT", "2222"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, s)

	log.Printf("order service is running on port %s\n", getEnv("ORDER_SERVER_PORT", "2222"))

	if err := grpcServer.Serve(lis); err != nil {
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

	mg, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}

	rb, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	db := repository.NewOrderRepo(mg)
	ps := rabbitmq.NewRabbitMQ(rb)

	s := NewOrder(db, ps)

	startGRPCServer(s)

}

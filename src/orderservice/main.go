package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	"github.com/pongsathonn/ihavefood/src/orderservice/internal"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	s := internal.NewOrderService(
		internal.NewOrderStorage(initMongoClient()),
		internal.NewRabbitMQ(initRabbitMQ()),
	)

	go s.StartConsume()

	startGRPCServer(s)
}

func initRabbitMQ() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("ORDER_AMQP_USER"),
		os.Getenv("ORDER_AMQP_PASS"),
		os.Getenv("ORDER_AMQP_HOST"),
	)
	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return conn
		}
		if i == maxRetries {
			log.Fatalf("Could not establish RabbitMQ connection after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(5 * time.Second)
	}

	log.Fatalf("Unexpected")
	return nil
}

func initMongoClient() *mongo.Client {

	uri := fmt.Sprintf("mongodb://%s:%s@%s/db?authSource=admin",
		os.Getenv("ORDER_DB_USER"),
		os.Getenv("ORDER_DB_PASS"),
		os.Getenv("ORDER_DB_HOST"),
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected to MongoDB")

	db := client.Database("db")

	if err := db.CreateCollection(context.TODO(), "orders"); err != nil {
		var alreayExistsColl mongo.CommandError
		if !errors.As(err, &alreayExistsColl) {
			log.Fatal("Failed to create collection:", err)
		}
	}

	coll := db.Collection("orders")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"requestId", 1}}, //preventing duplicate order
		Options: options.Index().SetUnique(true),
	}

	newIndex, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal("Failed to create index:", err)
	}

	slog.Info("created new mongo index", "name", newIndex)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping:", err)
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

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, s)

	log.Printf("order service is running on port %s\n", os.Getenv("ORDER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

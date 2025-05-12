package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	amqp "github.com/rabbitmq/amqp091-go"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
	"github.com/pongsathonn/ihavefood/src/restaurantservice/internal"
)

func main() {

	mongo, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}

	rabbitmq, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	s := internal.NewRestaurantService(
		internal.NewRestaurantStorage(mongo),
		internal.NewRabbitMQ(rabbitmq),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.RunMessageProcessing(ctx)

	startGRPCServer(s)
}

func initMongoClient() (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/restaurant_database?authSource=admin",
		os.Getenv("RESTAURANT_MONGO_USER"),
		os.Getenv("RESTAURANT_MONGO_PASS"),
		os.Getenv("RESTAURANT_MONGO_HOST"),
		os.Getenv("RESTAURANT_MONGO_PORT"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	coll := client.Database("restaurant_database", nil).Collection("restaurantCollection")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"restaurantName", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initRabbitMQ() (*amqp.Connection, error) {

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("RESTAURANT_AMQP_USER"),
		os.Getenv("RESTAURANT_AMQP_PASS"),
		os.Getenv("RESTAURANT_AMQP_HOST"),
		os.Getenv("RESTAURANT_AMQP_PORT"),
	))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func startGRPCServer(s *internal.RestaurantService) {

	uri := fmt.Sprintf(":%s", os.Getenv("RESTAURANT_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRestaurantServiceServer(grpcServer, s)

	log.Printf("restaurant service is running on port %s\n", os.Getenv("RESTAURANT_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

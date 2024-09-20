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

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
	"github.com/pongsathonn/ihavefood/src/restaurantservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	client, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	repository := internal.NewRestaurantRepository(client)
	rabbitmq := internal.NewRabbitMQ(conn)
	restaurantService := internal.NewRestaurantService(repository, rabbitmq)

	startGRPCServer(restaurantService)

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

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("RESTAURANT_AMQP_USER"),
		os.Getenv("RESTAURANT_AMQP_PASS"),
		os.Getenv("RESTAURANT_AMQP_HOST"),
		os.Getenv("RESTAURANT_AMQP_PORT"),
	)
	conn, err := amqp.Dial(uri)
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

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
	"github.com/pongsathonn/ihavefood/src/restaurantservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initMongoClient() (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/restaurant_database?authSource=admin",
		os.Getenv("RESTAURANT_MONGO_USER"),
		os.Getenv("RESTAURANT_MONGO_PASS"),
		os.Getenv("RESTAURANT_MONGO_HOST"),
		os.Getenv("RESTAURANT_MONGO_PORT"),
	)
	conn, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	coll := conn.Database("restaurant_database", nil).Collection("restaurantCollection")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"restaurantName", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return nil, err
	}
	return conn, nil
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

	uri := fmt.Sprintf(":%s", getEnv("RESTAURANT_SERVER_PORT", "1111"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRestaurantServiceServer(grpcServer, s)

	log.Printf("restaurant service is running on port %s\n", getEnv("RESTAURANT_SERVER_PORT", "1111"))

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

	mongoClient, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}

	rabbitConn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	repo := internal.NewRestaurantRepository(mongoClient)
	rb := internal.NewRabbitmqClient(rabbitConn)
	s := internal.NewRestaurantService(repo, rb)

	startGRPCServer(s)

}

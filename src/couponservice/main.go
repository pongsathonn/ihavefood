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

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	"github.com/pongsathonn/ihavefood/src/couponservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmq := initRabbitMQ()
	mongo := initMongoClient()

	repository := internal.NewCouponRepository(mongo)
	couponService := internal.NewCouponService(rabbitmq, repository)
	startGRPCServer(couponService)

}

func initRabbitMQ() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("COUPON_AMQP_USER", "donkadmin"),
		getEnv("COUPON_AMQP_PASS", "donkpassword"),
		getEnv("COUPON_AMQP_HOST", "localhost"),
		getEnv("COUPON_AMQP_PORT", "5672"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func initMongoClient() *mongo.Client {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/coupon_database?authSource=admin",
		os.Getenv("COUPON_MONGO_USER"),
		os.Getenv("COUPON_MONGO_PASS"),
		os.Getenv("COUPON_MONGO_HOST"),
		os.Getenv("COUPON_MONGO_PORT"),
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Database("coupon_database").CreateCollection(context.TODO(), "couponCollection")
	if err != nil {
		//TODO if exists pass
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	return client
}

func startGRPCServer(s *internal.CouponService) {

	uri := fmt.Sprintf(":%s", getEnv("COUPON_SERVER_PORT", "3333"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCouponServiceServer(grpcServer, s)

	log.Printf("coupon service is running on port %s\n", getEnv("COUPON_SERVER_PORT", "3333"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("%s doesn't exists \n", key)
	return defaultValue
}

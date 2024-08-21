package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	"github.com/pongsathonn/ihavefood/src/couponservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("COUPON_AMQP_USER", "donkadmin"),
		getEnv("COUPON_AMQP_PASS", "donkpassword"),
		getEnv("COUPON_AMQP_HOST", "localhost"),
		getEnv("COUPON_AMQP_PORT", "5672"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func initOrderClient() (pb.OrderServiceClient, error) {

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())

	port := getEnv("COUPON_SERVER_PORT", "3333")
	uri := fmt.Sprintf("localhost:%s", port)

	conn, err := grpc.NewClient(uri, opt)
	if err != nil {
		return nil, err
	}

	// defer conn.Close()

	return pb.NewOrderServiceClient(conn), nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(s *internal.CouponService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("COUPON_SERVER_PORT", "3333"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCouponServiceServer(grpcServer, s)

	log.Printf("coupon service is running on port %s\n", getEnv("COUPON_SERVER_PORT", "3333"))

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
	conn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	orderClient, err := initOrderClient()
	if err != nil {
		log.Fatal(err)
	}

	c := internal.NewCouponService(conn, orderClient)

	startGRPCServer(c)

	//-----------------------

}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pongsathonn/ihavefood/src/orderservice/internal"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
			}

			return a
		},
	}))
	slog.SetDefault(logger)
}

func main() {

	internal.SetupValidator()

	s := internal.NewOrderService(
		internal.NewOrderStorage(initMongoClient()),
		internal.NewRabbitMQ(initAMQPCon()),
		pb.NewCouponServiceClient(newGRPCConn("COUPON_URI")),
		pb.NewCustomerServiceClient(newGRPCConn("CUSTOMER_URI")),
		pb.NewDeliveryServiceClient(newGRPCConn("DELIVERY_URI")),
		pb.NewMerchantServiceClient(newGRPCConn("MERCHANT_URI")),
	)
	go s.StartConsume()

	uri := fmt.Sprintf(":%s", os.Getenv("ORDER_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, s)

	slog.Info("Order service started", "port", os.Getenv("ORDER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

func newGRPCConn(env string) *grpc.ClientConn {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv(env), opt)
	if err != nil {
		log.Fatalf("failed to create new grpc channel for %s: %v", env, err)
	}
	return conn
}

func initAMQPCon() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("RBMQ_ORDER_USER"),
		os.Getenv("RBMQ_ORDER_PASS"),
		os.Getenv("AMQP_SERVER_URI"),
	)
	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			slog.Info("AMQP connection established")
			return conn
		}
		if i == maxRetries {
			log.Fatalf("Could not establish AMQP connection after %d attempts: %v", maxRetries, err)
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

	slog.Info("MongoDB connection established")

	db := client.Database("db")

	if err := db.CreateCollection(context.TODO(), "orders"); err != nil {
		var alreayExistsColl mongo.CommandError
		if !errors.As(err, &alreayExistsColl) {
			log.Fatal("Failed to create collection:", err)
		}
	}

	coll := db.Collection("orders")
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "requestId", Value: 1}, // preventing duplicate order
		},
		Options: options.Index().SetUnique(true),
	}

	newIndex, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal("Failed to create index:", err)
	}

	slog.Info("MongoDB index created", "index", newIndex)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping:", err)
	}

	return client

}

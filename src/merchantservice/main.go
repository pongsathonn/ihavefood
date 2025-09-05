package main

import (
	"context"
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

	amqp "github.com/rabbitmq/amqp091-go"

	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
	"github.com/pongsathonn/ihavefood/src/merchantservice/internal"
)

func main() {
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

	mongo, err := initMongoClient()
	if err != nil {
		log.Fatal(err)
	}

	s := internal.NewMerchantService(
		internal.NewMerchantStorage(mongo),
		internal.NewRabbitMQ(initAMQPCon()),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.RunMessageProcessing(ctx)

	startGRPCServer(s)
}

func initMongoClient() (*mongo.Client, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s/db?authSource=admin",
		os.Getenv("MERCHANT_DB_USER"),
		os.Getenv("MERCHANT_DB_PASS"),
		os.Getenv("MERCHANT_DB_HOST"),
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

	coll := client.Database("db", nil).Collection("merchants")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"name", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initAMQPCon() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("RBMQ_MERCHANT_USER"),
		os.Getenv("RBMQ_MERCHANT_PASS"),
		os.Getenv("AMQP_SERVER_URI"),
	)
	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			slog.Info("Successfully connected to AMQP Server")
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

func startGRPCServer(s *internal.MerchantService) {

	uri := fmt.Sprintf(":%s", os.Getenv("MERCHANT_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMerchantServiceServer(grpcServer, s)

	slog.Info("merchant service is running", "port", os.Getenv("MERCHANT_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

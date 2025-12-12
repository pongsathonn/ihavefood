package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pongsathonn/ihavefood/src/merchantservice/internal"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"google.golang.org/grpc"

	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initAMQPCon() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s/%s",
		os.Getenv("RBMQ_USER"),
		os.Getenv("RBMQ_PASS"),
		os.Getenv("RBMQ_HOST"),
		os.Getenv("RBMQ_USER"),
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

func initMongoDB() *mongo.Collection {

	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?appName=%s",
		url.QueryEscape(os.Getenv("MONGO_USER")),
		url.QueryEscape(os.Getenv("MONGO_PASS")),
		url.QueryEscape(os.Getenv("MONGO_HOST")),
		url.QueryEscape(os.Getenv("MONGO_CLUSTER")),
	)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI).SetTimeout(30 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	coll := client.Database("merchantdb", nil).Collection("merchants")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	return coll
}

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

	rabbitmq := internal.NewRabbitMQ(initAMQPCon())
	srv := internal.NewMerchantService(
		internal.NewMerchantStorage(initMongoDB()),
		rabbitmq,
	)

	go rabbitmq.Start([]*internal.EventHandler{
		{
			Queue:   "merchant_assign_queue",
			Key:     "order.placed.event",
			Handler: srv.HandlePlaceOrder,
		},
	})

	uri := fmt.Sprintf(":%s", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	pb.RegisterMerchantServiceServer(grpcServer, srv)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"time"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	"github.com/pongsathonn/ihavefood/src/orderservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
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
		internal.NewOrderStorage(initMongoDB()),
		internal.NewRabbitMQ(initAMQPCon()),
		pb.NewCouponServiceClient(newGRPCConn("COUPON_URI")),
		pb.NewCustomerServiceClient(newGRPCConn("CUSTOMER_URI")),
		pb.NewDeliveryServiceClient(newGRPCConn("DELIVERY_URI")),
		pb.NewMerchantServiceClient(newGRPCConn("MERCHANT_URI")),
	)
	go s.StartConsume()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("net.Listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)
	pb.RegisterOrderServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func newGRPCConn(env string) *grpc.ClientConn {

	ctx := context.Background()

	uri := os.Getenv(env)
	tokenSource, err := idtoken.NewTokenSource(ctx, uri)
	if err != nil {
		log.Fatalf("Failed to create token source: %v", err)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("Failed to parse URI: %v", err)
	}

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal("failed to load system root CA cert pool:", err)
	}

	var creds grpc.DialOption
	if env != "DELIVERY_URI" {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	} else {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		}))
	}

	cc, err := grpc.NewClient(
		parsedURL.Host+":443",
		creds,
		grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: tokenSource}),
	)
	if err != nil {
		log.Fatalf("failed to create grpc client for %s: %v", env, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	healthClient := healthgrpc.NewHealthClient(cc)
	resp, err := healthClient.Check(ctx, &healthgrpc.HealthCheckRequest{
		Service: "",
	})
	if err != nil {
		log.Fatalf("health check failed for %s: %v", env, err)
	}

	if resp.Status != healthgrpc.HealthCheckResponse_SERVING {
		log.Fatalf("service not healthy for %s: %v", env, resp.Status)
	}

	return cc
}

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
			slog.Info("Successfully connected to AMQP server")
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

	coll := client.Database("orderdb", nil).Collection("orders")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "requestId", Value: 1}, // preventing duplicate order
		},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	return coll
}

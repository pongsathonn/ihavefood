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

	"google.golang.org/grpc"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	"github.com/pongsathonn/ihavefood/src/couponservice/internal"
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

	// indexModel := mongo.IndexModel{
	// 	Keys:    bson.D{{Key: "name", Value: 1}},
	// 	Options: options.Index().SetUnique(true),
	// }
	// _, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	// if err != nil {
	// 	panic(err)
	// }

	return client.Database("coupondb", nil).Collection("coupons")
}

func startGRPCServer(s *internal.CouponService) {

	uri := fmt.Sprintf(":%s", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)

	pb.RegisterCouponServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// cleanUpCoupons runs a scheduled job that removes expired coupons
// or coupons with zero quantity from the database. It executes this
// cleanup operation every 30 minutes.
func cleanUpCoupons(ctx context.Context, coll *mongo.Collection) {

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiredTime := time.Now().Unix()
			filter := bson.M{"$or": []bson.M{
				{"expiration": bson.M{"$lt": expiredTime}},
				{"quantity": bson.M{"$lt": 1}},
			}}
			if _, err := coll.DeleteMany(ctx, filter); err != nil {
				slog.Error("clean up coupons failed", "err", err)
			}
		case <-ctx.Done():
			slog.Info("clean up stopped")
			return
		}
	}
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

	mongo := initMongoDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cleanUpCoupons(ctx, mongo)

	startGRPCServer(internal.NewCouponService(
		initAMQPCon(),
		internal.NewCouponStorage(mongo),
	))
}

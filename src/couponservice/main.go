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

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	"github.com/pongsathonn/ihavefood/src/couponservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initAMQPCon() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("RBMQ_COUPON_USER"),
		os.Getenv("RBMQ_COUPON_PASS"),
		os.Getenv("AMQP_SERVER_URI"),
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

func initMongoClient() *mongo.Client {

	uri := fmt.Sprintf("mongodb://%s:%s@%s/db?authSource=admin",
		os.Getenv("COUPON_DB_USER"),
		os.Getenv("COUPON_DB_PASS"),
		os.Getenv("COUPON_DB_HOST"),
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Database("db").CreateCollection(context.TODO(), "coupons")
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

	uri := fmt.Sprintf(":%s", os.Getenv("COUPON_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCouponServiceServer(grpcServer, s)

	slog.Info("coupon service is running", "port", os.Getenv("COUPON_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// cleanUpCoupons runs a scheduled job that removes expired coupons
// or coupons with zero quantity from the database. It executes this
// cleanup operation every 30 minutes.
//
// Parameters:
//   - ctx: A context to allow for graceful shutdown of the job.
//   - client: A MongoDB client used to interact with the database.
//
// The function runs in an infinite loop and performs the following:
//   - Removes any coupon whose 'expiration' field is less than the current time.
//   - Removes any coupon whose 'quantity' field is less than 1.
//   - Stops when the context is canceled, allowing for graceful termination.
func cleanUpCoupons(ctx context.Context, client *mongo.Client) {

	coll := client.Database("db", nil).Collection("coupons")

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

	mongo := initMongoClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cleanUpCoupons(ctx, mongo)

	startGRPCServer(internal.NewCouponService(
		initAMQPCon(),
		internal.NewCouponStorage(mongo),
	))
}

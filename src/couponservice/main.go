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

	pb "github.com/pongsathonn/ihavefood/src/couponservice/genproto"
	"github.com/pongsathonn/ihavefood/src/couponservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitmq := initRabbitMQ()
	mongo := initMongoClient()

	repository := internal.NewCouponRepository(mongo)
	couponService := internal.NewCouponService(rabbitmq, repository)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanUpCoupons(ctx, mongo)

	startGRPCServer(couponService)
}

func initRabbitMQ() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("COUPON_AMQP_USER"),
		os.Getenv("COUPON_AMQP_PASS"),
		os.Getenv("COUPON_AMQP_HOST"),
		os.Getenv("COUPON_AMQP_PORT"),
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

	uri := fmt.Sprintf(":%s", os.Getenv("COUPON_SERVER_PORT", "3333"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCouponServiceServer(grpcServer, s)

	log.Printf("coupon service is running on port %s\n", os.Getenv("COUPON_SERVER_PORT", "3333"))

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

	coll := client.Database("coupon_database", nil).Collection("couponCollection")

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
				log.Printf("Clean Up delete failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("Clean Up stopped")
			return
		}
	}
}

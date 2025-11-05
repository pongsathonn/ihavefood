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

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// //////////////////////////////////////////////////////////////////////////
type FakeRabbitMQ struct{}

func (f *FakeRabbitMQ) Publish(ctx context.Context, routingKey string, msg amqp.Publishing) error {
	// Just log instead of sending
	fmt.Printf("FakeRabbitMQ.Publish called: routingKey=%s, body=%s\n", routingKey, string(msg.Body))
	return nil
}

func (f *FakeRabbitMQ) Subscribe(ctx context.Context, queue, routingkey string) (<-chan amqp.Delivery, error) {
	fmt.Printf("FakeRabbitMQ.Subscribe called: queue=%s, routingKey=%s\n", queue, routingkey)
	ch := make(chan amqp.Delivery)
	close(ch) // immediately close so consuming goroutines don’t block
	return ch, nil
}

////////////////////////////////////////////////////////////////////////////

// getSecret panics because it tends to use in main only.
func getSecret(secretName string) string {

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		panic("GCP_PROJECT_ID environment variable is not set")
	}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	versionName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest",
		projectID,
		secretName,
	)

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: versionName,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		panic(err)
	}

	return string(result.Payload.Data)
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

func initMongoDB() *mongo.Client {

	uri := fmt.Sprintf("mongodb+srv://%s:%s@mymongodb.jtcdxwq.mongodb.net/?appName=MyMongoDB",
		url.QueryEscape(getSecret("MERCHANT_DB_USER")),
		url.QueryEscape(getSecret("MERCHANT_DB_PASS")),
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

	coll := client.Database("db", nil).Collection("merchants")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	return client

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

	srv := internal.NewMerchantService(
		internal.NewMerchantStorage(initMongoDB()),
		&FakeRabbitMQ{},
		// internal.NewRabbitMQ(initAMQPCon()),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.RunMessageProcessing(ctx)

	uri := fmt.Sprintf(":%s", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMerchantServiceServer(grpcServer, srv)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

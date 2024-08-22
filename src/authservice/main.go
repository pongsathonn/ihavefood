package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/lib/pq"
	"github.com/pongsathonn/ihavefood/src/authservice/internal"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := internal.InitSigningKey(); err != nil {
		log.Printf("Failed to initialize jwt signingkey:", err)
	}

	// Initialize dependencies
	db, err := initPostgres()
	if err != nil {
		log.Fatal("Failed to initialize PostgresDB connection:", err)
	}

	amqpConn, err := initRabbitMQ()
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ connection:", err)
	}
	rb := internal.NewRabbitMQ(amqpConn)

	userClient, err := newUserServiceClient()
	if err != nil {
		log.Fatal("Failed to make new user client:", err)
	}

	auth := internal.NewAuthService(db, rb, userClient)
	startGRPCServer(auth)

}

func newUserServiceClient() (pb.UserServiceClient, error) {

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(getEnv("USER_URI", ""), opt)
	if err != nil {
		return nil, err
	}
	client := pb.NewUserServiceClient(conn)

	return client, nil
}

func initPostgres() (*sql.DB, error) {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/auth_database?sslmode=disable",
		getEnv("AUTH_POSTGRES_USER", "donkadmin"),
		getEnv("AUTH_POSTGRES_PASS", "donkpassword"),
		getEnv("AUTH_POSTGRES_HOST", "localhost"),
		getEnv("AUTH_POSTGRES_PORT", "5432"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("AUTH_AMQP_USER", "donkadmin"),
		getEnv("AUTH_AMQP_PASS", "donkpassword"),
		getEnv("AUTH_AMQP_HOST", "localhost"),
		getEnv("AUTH_AMQP_PORT", "5672"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to dial AMQP connection: %v", err)
	}

	return conn, nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(s *internal.AuthService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("AUTH_SERVER_PORT", "4444"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	log.Printf("auth service is running on port %s\n", getEnv("AUTH_SERVER_PORT", "4444"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}

}

// getEnv fetches an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {

	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	log.Printf("%s doesn't exists \n", key)
	return defaultValue
}

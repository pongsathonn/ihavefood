package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pongsathonn/ihavefood/src/authservice/internal"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	if err := internal.InitSigningKey(); err != nil {
		slog.Error("initilize jwt signing key", "err", err)
	}

	db, err := initPostgres()
	if err != nil {
		log.Fatalf("Failed to initialize PostgresDB connection: %v", err)
	}
	storage := internal.NewAuthStorage(db)

	if err := internal.InitAdminUser(storage); err != nil {
		log.Printf("Failed to initialize create admin user: %v", err)
	}

	amqpConn, err := initRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ connection: %v", err)
	}

	userClient, err := newProfileServiceClient()
	if err != nil {
		log.Fatalf("Failed to initialize ProfileService connection: %v", err)
	}

	authService := internal.NewAuthService(
		storage,
		internal.NewRabbitMQ(amqpConn),
		userClient,
	)
	startGRPCServer(authService)

}

func newProfileServiceClient() (pb.ProfileServiceClient, error) {

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv("USER_URI"), opt)
	if err != nil {
		return nil, err
	}
	client := pb.NewProfileServiceClient(conn)

	return client, nil
}

func initPostgres() (*sql.DB, error) {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/auth_database?sslmode=disable",
		os.Getenv("AUTH_POSTGRES_USER"),
		os.Getenv("AUTH_POSTGRES_PASS"),
		os.Getenv("AUTH_POSTGRES_HOST"),
		os.Getenv("AUTH_POSTGRES_PORT"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func initRabbitMQ() (*amqp.Connection, error) {

	const maxRetries = 5
	var errDialRabbitmq error

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("AUTH_AMQP_USER"),
		os.Getenv("AUTH_AMQP_PASS"),
		os.Getenv("AUTH_AMQP_HOST"),
		os.Getenv("AUTH_AMQP_PORT"),
	)

	for _ = range maxRetries {
		conn, err := amqp.Dial(uri)
		if err == nil {
			return conn, nil
		}
		errDialRabbitmq = err
	}

	return nil, errDialRabbitmq
}

// startGRPCServer sets up and starts the gRPC server
// func startGRPCServer(s *internal.AuthService) {
func startGRPCServer(s *internal.AuthService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", os.Getenv("AUTH_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	log.Printf("auth service is running on port %s\n", os.Getenv("AUTH_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

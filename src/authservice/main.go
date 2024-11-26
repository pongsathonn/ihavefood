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

	conn, err := grpc.NewClient(os.Getenv("PROFILE_URI"), opt)
	if err != nil {
		return nil, err
	}

	slog.Info("Channel for ProfileServiceClient created successfully")
	return pb.NewProfileServiceClient(conn), nil
}

func initPostgres() (*sql.DB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbName := "postgres"
	host := os.Getenv("AUTH_POSTGRES_HOST")
	port := os.Getenv("AUTH_POSTGRES_PORT")

	db, err := sql.Open(dbName, fmt.Sprintf("postgres://%s:%s@%s:%s/auth_database?sslmode=disable",
		os.Getenv("AUTH_POSTGRES_USER"),
		os.Getenv("AUTH_POSTGRES_PASS"),
		host,
		port,
	))
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	slog.Info("Database initialized successfully",
		"db", dbName,
		"host", host,
		"port", port,
	)

	return db, nil
}

func initRabbitMQ() (*amqp.Connection, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("AUTH_AMQP_USER"),
		os.Getenv("AUTH_AMQP_PASS"),
		os.Getenv("AUTH_AMQP_HOST"),
		os.Getenv("AUTH_AMQP_PORT"),
	)

	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		conn, err := amqp.Dial(uri)
		if err == nil {
			slog.Info("RabbitMQ initialization successful",
				"host", os.Getenv("AUTH_AMQP_HOST"),
				"port", os.Getenv("AUTH_AMQP_PORT"),
			)
			return conn, nil
		}

		slog.Warn(fmt.Sprintf("Failed to connect to RabbitMQ, retrying... (%d/5)", i+1))
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts", maxRetries)
}

// startGRPCServer sets up and starts the gRPC server
// func startGRPCServer(s *internal.AuthService) {
func startGRPCServer(auths *internal.AuthService) {

	uri := fmt.Sprintf(":%s", os.Getenv("AUTH_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, auths)

	slog.Info("AuthService initialized successfully",
		"port", os.Getenv("AUTH_SERVER_PORT"),
	)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

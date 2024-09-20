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

	"github.com/pongsathonn/ihavefood/src/authservice/internal"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	if err := internal.InitSigningKey(); err != nil {
		log.Printf("Failed to initialize jwt signingkey:", err)
	}

	db, err := initPostgres()
	if err != nil {
		log.Fatal("Failed to initialize PostgresDB connection:", err)
	}
	if err := internal.InitAdminUser(db); err != nil {
		log.Printf("Failed to initialize create admin user:", err)
	}

	amqpConn, err := initRabbitMQ()
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ connection:", err)
	}
	rabbitmq := internal.NewRabbitMQ(amqpConn)

	userClient, err := newUserServiceClient()
	if err != nil {
		log.Fatal("Failed to make new user client:", err)
	}

	authService := internal.NewAuthService(db, rabbitmq, userClient)
	startGRPCServer(authService)

}

func newUserServiceClient() (pb.UserServiceClient, error) {

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv("USER_URI"), opt)
	if err != nil {
		return nil, err
	}
	client := pb.NewUserServiceClient(conn)

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
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("AUTH_AMQP_USER"),
		os.Getenv("AUTH_AMQP_PASS"),
		os.Getenv("AUTH_AMQP_HOST"),
		os.Getenv("AUTH_AMQP_PORT"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to dial AMQP connection: %v", err)
	}

	return conn, nil
}

// startGRPCServer sets up and starts the gRPC server
// func startGRPCServer(s *internal.AuthService) {
func startGRPCServer(s *internal.AuthService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", os.Getenv("AUTH_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	log.Printf("auth service is running on port %s\n", os.Getenv("AUTH_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}

}

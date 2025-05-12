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
)

func main() {

	if err := initTimeZone(); err != nil {
		slog.Error("failed to init time zone", "err", err)
	}

	if err := internal.InitSigningKey(); err != nil {
		slog.Error("failed to init jwt signing key", "err", err)
	}

	db, err := initPostgres()
	if err != nil {
		log.Fatalf("Failed to initialize PostgresDB connection: %v", err)
	}
	storage := internal.NewStorage(db)

	if err := internal.CreateAdmin(storage); err != nil {
		log.Printf("Failed to create admin user: %v", err)
	}

	userClient, err := newProfileServiceClient()
	if err != nil {
		log.Fatalf("Failed to initialize ProfileService connection: %v", err)
	}

	service := internal.NewAuthService(storage, userClient)
	startGRPCServer(service)

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

	host := os.Getenv("AUTH_POSTGRES_HOST")
	port := os.Getenv("AUTH_POSTGRES_PORT")

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/auth_database?sslmode=disable",
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
		"host", host,
		"port", port,
	)

	return db, nil
}

// startGRPCServer sets up and starts the gRPC server
// func startGRPCServer(s *internal.AuthService) {
func startGRPCServer(s *internal.AuthService) {

	uri := fmt.Sprintf(":%s", os.Getenv("AUTH_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	slog.Info("AuthService initialized successfully",
		"port", os.Getenv("AUTH_SERVER_PORT"),
	)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func initTimeZone() error {
	l, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return err
	}

	time.Local = l
	return nil
}

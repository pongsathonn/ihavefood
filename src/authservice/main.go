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

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var signingKey []byte

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
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil

}

func initSigningKey() error {

	key := os.Getenv("JWT_SIGNING_KEY")

	if key == "" {
		return fmt.Errorf("JWT_SIGNING_KEY environment variable is empty")
	}

	signingKey = []byte(key)

	return nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(a *authService) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("AUTH_SERVER_PORT", "4444"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	// Create and start the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, a)

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
	return defaultValue
}

func main() {

	// Initialize dependencies
	db, err := initPostgres()
	if err != nil {
		log.Fatal("Failed to initialize PostgresDB connection:", err)
	}

	auth := NewAuthService(db)

	startGRPCServer(auth)

}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	"github.com/pongsathonn/ihavefood/src/authservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initPostgres() (*sql.DB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user := os.Getenv("AUTH_DB_USER")
	pass := os.Getenv("AUTH_DB_PASS")
	host := os.Getenv("AUTH_DB_HOST")
	dbName := os.Getenv("AUTH_DB_NAME")

	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		user, pass, host, dbName,
	))
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	slog.Info("Database initialized successfully", "host", host)
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

func initAMQPCon() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("RBMQ_AUTH_USER"),
		os.Getenv("RBMQ_AUTH_PASS"),
		os.Getenv("AMQP_SERVER_URI"),
	)
	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			slog.Info("Successfully connected to AMQP")
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

	internal.SetupValidator()

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

	auth := internal.NewAuthService(
		internal.NewStorage(db),
		internal.NewRabbitMQ(initAMQPCon()),
	)

	if err := auth.CreateDemoUsers(); err != nil {
		slog.Error("Failed to create Demo Users", "err", err)
	}

	startGRPCServer(auth)
}

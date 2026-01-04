package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"path/filepath"

	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/pongsathonn/ihavefood/src/customerservice/genproto"
	"github.com/pongsathonn/ihavefood/src/customerservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func dbPool() (*pgxpool.Pool, error) {
	dsn := os.Getenv("CUSTOMER_DB_URL")
	if dsn == "" {
		return nil, fmt.Errorf("CUSTOMER_DB_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
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
	pool, err := dbPool()
	if err != nil {
		log.Fatal(err)
	}

	rabbitmq := internal.NewRabbitMQ(initAMQPCon())
	s := internal.NewCustomerService(rabbitmq,
		internal.NewCustomerStorage(pool),
	)

	go rabbitmq.Start([]*internal.EventHandler{
		{Key: "sync.customer.created", Handler: s.HandleCustomerCreation},
	})

	if s == nil {
		log.Fatal("customer service instance is nil")
	}

	uri := fmt.Sprintf(":%s", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServer, healthcheck)
	pb.RegisterCustomerServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}

}

func initAMQPCon() *amqp.Connection {

	uri := fmt.Sprintf("amqp://%s:%s@%s/%s",
		os.Getenv("RBMQ_USER"),
		os.Getenv("RBMQ_PASS"),
		os.Getenv("RBMQ_HOST"),
		os.Getenv("RBMQ_USER"),
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

package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"path/filepath"

	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/customerservice/genproto"
	"github.com/pongsathonn/ihavefood/src/customerservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	db, err := initPostgres()
	if err != nil {
		log.Fatal(err)
	}

	customerService := internal.NewCustomerService(
		internal.NewRabbitMQ(initAMQPCon()),
		internal.NewCustomerStorage(db),
	)

	startGRPCServer(customerService)

}

func initAMQPCon() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("RBMQ_CUSTOMER_USER"),
		os.Getenv("RBMQ_CUSTOMER_PASS"),
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

func initPostgres() (*sql.DB, error) {

	uri := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("CUSTOMER_DB_USER"),
		os.Getenv("CUSTOMER_DB_PASS"),
		os.Getenv("CUSTOMER_DB_HOST"),
		os.Getenv("CUSTOMER_DB_NAME"),
	)

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil

}

func startGRPCServer(s *internal.CustomerService) {

	if s == nil {
		log.Fatal("customer service instance is nil")
	}

	uri := fmt.Sprintf(":%s", os.Getenv("CUSTOMER_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCustomerServiceServer(grpcServer, s)

	slog.Info("customer service is running", "port", os.Getenv("CUSTOMER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

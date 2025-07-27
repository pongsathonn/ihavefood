package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/customerservice/genproto"
	"github.com/pongsathonn/ihavefood/src/customerservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	db, err := initPostgres()
	if err != nil {
		log.Fatal(err)
	}

	customerService := internal.NewCustomerService(
		internal.NewRabbitMQ(initRabbitMQ()),
		internal.NewCustomerStorage(db),
	)

	startGRPCServer(customerService)

}

func initRabbitMQ() *amqp.Connection {
	uri := fmt.Sprintf("amqp://%s:%s@%s",
		os.Getenv("CUSTOMER_AMQP_USER"),
		os.Getenv("CUSTOMER_AMQP_PASS"),
		os.Getenv("CUSTOMER_AMQP_HOST"),
	)
	maxRetries := 5
	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return conn
		}

		if i == maxRetries {
			log.Fatalf("Could not establish RabbitMQ connection after %d attempts: %v", maxRetries, err)
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
		log.Fatal("profile service instance is nil")
	}

	uri := fmt.Sprintf(":%s", os.Getenv("CUSTOMER_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCustomerServiceServer(grpcServer, s)

	log.Printf("profile service is running on port %s\n", os.Getenv("CUSTOMER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

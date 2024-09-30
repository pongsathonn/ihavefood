package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
	"github.com/pongsathonn/ihavefood/src/userservice/internal"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	db, err := initPostgres()
	if err != nil {
		log.Fatal(err)
	}

	amqpConn, err := initRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}

	repository := internal.NewUserRepository(db)
	rabbitmq := internal.NewRabbitMQ(amqpConn)
	userService := internal.NewUserService(rabbitmq, repository)
	startGRPCServer(userService)

}

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("USER_AMQP_USER"),
		os.Getenv("USER_AMQP_PASS"),
		os.Getenv("USER_AMQP_HOST"),
		os.Getenv("USER_AMQP_PORT"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func initPostgres() (*sql.DB, error) {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/user_database?sslmode=disable",
		os.Getenv("USER_POSTGRES_USER"),
		os.Getenv("USER_POSTGRES_PASS"),
		os.Getenv("USER_POSTGRES_HOST"),
		os.Getenv("USER_POSTGRES_PORT"),
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

func startGRPCServer(s *internal.UserService) {

	if s == nil {
		log.Fatal("userService instance is nil")
	}

	uri := fmt.Sprintf(":%s", os.Getenv("USER_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, s)

	log.Printf("user service is running on port %s\n", os.Getenv("USER_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

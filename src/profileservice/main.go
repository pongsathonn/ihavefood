package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/pongsathonn/ihavefood/src/profileservice/genproto"
	"github.com/pongsathonn/ihavefood/src/profileservice/internal"
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

	profileService := internal.NewProfileService(
		internal.NewRabbitMQ(amqpConn),
		internal.NewProfileStorage(db),
	)

	startGRPCServer(profileService)

}

// initPubSub initializes the RabbitMQ connection and returns the rabbitmq instance
func initRabbitMQ() (*amqp.Connection, error) {

	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		os.Getenv("PROFILE_AMQP_USER"),
		os.Getenv("PROFILE_AMQP_PASS"),
		os.Getenv("PROFILE_AMQP_HOST"),
		os.Getenv("PROFILE_AMQP_PORT"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func initPostgres() (*sql.DB, error) {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/profile_database?sslmode=disable",
		os.Getenv("PROFILE_POSTGRES_USER"),
		os.Getenv("PROFILE_POSTGRES_PASS"),
		os.Getenv("PROFILE_POSTGRES_HOST"),
		os.Getenv("PROFILE_POSTGRES_PORT"),
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

func startGRPCServer(s *internal.ProfileService) {

	if s == nil {
		log.Fatal("profile service instance is nil")
	}

	uri := fmt.Sprintf(":%s", os.Getenv("PROFILE_SERVER_PORT"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProfileServiceServer(grpcServer, s)

	log.Printf("profile service is running on port %s\n", os.Getenv("PROFILE_SERVER_PORT"))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

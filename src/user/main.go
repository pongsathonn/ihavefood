package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/food-delivery/src/user/genproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	pb.UnimplementedUserServiceServer

	db *sql.DB
}

func NewUserService(db *sql.DB) *userService {
	return &userService{db: db}
}

func (s *userService) UpdateUser(context.Context, *pb.Empty) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}

func (s *userService) CreateUser(context.Context, *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}

func (s *userService) ListUser(context.Context, *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
}

func (s *userService) GetUser(context.Context, *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

func (s *userService) DeleteUser(context.Context, *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}

func initPostgres() *sql.DB {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/user_database?sslmode=disable",
		os.Getenv("USER_POSTGRES_USER"),
		os.Getenv("USER_POSTGRES_PASS"),
		os.Getenv("USER_POSTGRES_HOST"),
		os.Getenv("USER_POSTGRES_PORT"),
	)

	db, err := sql.Open("postgres", uri)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db

}

func main() {
	db := initPostgres()

	userService := NewUserService(db)

	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(grpcServer, userService)

	port := os.Getenv("USER_SERVER_PORT")
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("User server is running on port %s", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

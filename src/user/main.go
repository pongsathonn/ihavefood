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
	"golang.org/x/crypto/bcrypt"
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

func (s *userService) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoginUser not implemented")
}

func (s *userService) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*pb.Empty, error) {

	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "failed xxx")
	}

	//TODO email validate ( regex ) , check username email exists, logging error

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed xxx")
	}

	_, err = s.db.Exec("INSERT INTO user_table(username, email, password) VALUES($1, $2, $3)",
		in.Username,
		in.Email,
		string(hashedPass),
	)

	if err != nil {
		log.Println("error create user :", err)
		return nil, status.Errorf(codes.Internal, "failed xxx")
	}

	return nil, nil
}

func (s *userService) ListUser(ctx context.Context, in *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
}
func (s *userService) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (s *userService) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}

func initPostgres() *sql.DB {

	// for development
	//connStr := fmt.Sprintf("postgres://donkadmin:donkpassword@localhost:5432/user_database?sslmode=disable")

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

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/pongsathonn/food-delivery/src/user/genproto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
	Email    string `bson:"email"`
}

func (us *userService) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.Empty, error) {

	return nil, nil
}

func (us *userService) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*pb.Empty, error) {

	if in.Username == "" || in.Password == "" || in.Email == "" {
		return nil, fmt.Errorf("request shoundn't empty")

	}

	//TODO hashing password bcrypt

	u := User{Username: in.Username, Password: in.Password, Email: in.Email}

	println("Hi")
	coll := us.conn.Database("user_Database", nil).Collection("userCollection")
	println("Hello")

	user_id, err := coll.InsertOne(context.TODO(), &u)
	if err != nil {
		return nil, err
	}

	log.Println("create user success user_id = ", user_id)

	return &pb.Empty{}, nil
}

func (us *userService) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.User, error) {
	log.Println("Print this == UserService Ja :", in.Username)
	return &pb.User{Username: "cojohn", Email: "john@mail.com", PhoneNumber: "09123", Address: nil}, nil
}
func (us *userService) ListUser(ctx context.Context, in *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	return nil, nil
}
func (us *userService) ForgotUserPassword(ctx context.Context, in *pb.ForgotUserPasswordRequest) (*pb.Empty, error) {
	return nil, nil
}
func (us *userService) ChangeUsername(ctx context.Context, in *pb.ChangeUsernameRequest) (*pb.Empty, error) {
	log.Println(in.Username)
	log.Println(in.NewUsername)
	return nil, nil
}
func (us *userService) ChangeEmail(ctx context.Context, in *pb.ChangeEmailRequest) (*pb.Empty, error) {
	return nil, nil
}
func (us *userService) ChangePassword(ctx context.Context, in *pb.ChangePasswordRequest) (*pb.Empty, error) {
	return nil, nil
}
func (us *userService) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, nil
}

type userService struct {
	pb.UnimplementedUserServiceServer
	conn *mongo.Client
}

func NewUserService(conn *mongo.Client) *userService {
	return &userService{conn: conn}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("USER_DB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	var us *userService
	us = NewUserService(client)

	userUri := os.Getenv("USER_URI")

	lis, err := net.Listen("tcp", userUri)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, us)

	log.Printf("server running on port %s\n", userUri)
	log.Fatalln(s.Serve(lis))

}

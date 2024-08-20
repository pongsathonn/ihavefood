package internal

import (
	"context"
	"database/sql"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

type userProfile struct {
	UserId      string
	Username    string
	Email       string
	PhoneNumber string
	AddressName sql.NullString
	SubDistrict sql.NullString
	District    sql.NullString
	Province    sql.NullString
	PostalCode  sql.NullString
}

// userService handle user profiles
type UserService struct {
	pb.UnimplementedUserServiceServer

	rabbitmq   RabbitmqClient
	repository UserRepository
}

func NewUserService(rabbitmq RabbitmqClient, repo UserRepository) *UserService {
	return &UserService{rabbitmq: rabbitmq, repository: repo}
}

func (x *UserService) UpdateUserProfile(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {

	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}

// this function will subscribe to AuthService for New User Register
// and save it to database
func (x *UserService) CreateUserProfile(ctx context.Context, in *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error) {

	if in.Username == "" || in.Email == "" || in.PhoneNumber == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "username, email, phone number or address must be provided")
	}

	userID, err := x.repository.SaveUserProfile(ctx, in.Username, in.Email, in.PhoneNumber, in.Address)
	if err != nil {
		log.Println("insert failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to save user to database")
	}

	return &pb.CreateUserProfileResponse{UserId: userID}, nil
}

func (x *UserService) ListUserProfile(ctx context.Context, req *pb.ListUserProfileRequest) (*pb.ListUserProfileResponse, error) {

	//TODO validate input

	return &pb.ListUserProfileResponse{UserProfiles: nil}, nil
}

func (x *UserService) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {

	return &pb.GetUserProfileResponse{UserProfile: nil}, nil
}

func (x *UserService) DeleteUserProfile(context.Context, *pb.DeleteUserProfileRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

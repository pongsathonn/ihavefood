package internal

import (
	"context"
	"database/sql"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

//TODO be careful about return error from database , it might contians sensitive information

// UserService handle UserProfile
type UserService struct {
	pb.UnimplementedUserServiceServer

	rabbitmq   RabbitMQ
	repository UserRepository
}

func NewUserService(rabbitmq RabbitMQ, repo UserRepository) *UserService {
	return &UserService{
		rabbitmq:   rabbitmq,
		repository: repo,
	}
}

func (x *UserService) CreateUserProfile(ctx context.Context, in *pb.CreateUserProfileRequest) (*pb.CreateUserProfileResponse, error) {

	if in.Username == "" || in.PhoneNumber == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "username,  phone number or address must be provided")
	}

	address := &address{
		addressName: sql.NullString{String: in.Address.AddressName, Valid: true},
		subDistrict: sql.NullString{String: in.Address.SubDistrict, Valid: true},
		district:    sql.NullString{String: in.Address.District, Valid: true},
		province:    sql.NullString{String: in.Address.Province, Valid: true},
		postalCode:  sql.NullString{String: in.Address.PostalCode, Valid: true},
	}
	userID, err := x.repository.SaveUserProfile(ctx, in.Username, in.PhoneNumber, address)
	if err != nil {
		log.Println("insert failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to save user to database")
	}

	return &pb.CreateUserProfileResponse{UserId: userID}, nil
}

func (x *UserService) GetUserProfileByUsername(ctx context.Context, in *pb.GetUserProfileByUsernameRequest) (*pb.GetUserProfileByUsernameResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	user, err := x.repository.UserProfile(ctx, in.Username)
	if err != nil {
		return nil, err
	}

	userProfile := &pb.UserProfile{
		UserId:      user.userId,
		Username:    user.username,
		PhoneNumber: user.phoneNumber,
		Address: &pb.Address{
			AddressName: user.address.addressName.String,
			SubDistrict: user.address.subDistrict.String,
			District:    user.address.district.String,
			Province:    user.address.province.String,
			PostalCode:  user.address.postalCode.String,
		},
	}
	return &pb.GetUserProfileByUsernameResponse{UserProfile: userProfile}, nil
}

func (x *UserService) ListUserProfile(ctx context.Context, in *pb.ListUserProfileRequest) (*pb.ListUserProfileResponse, error) {

	users, err := x.repository.UserProfiles(ctx)
	if err != nil {
		return nil, err
	}

	var userProfiles []*pb.UserProfile
	for _, user := range users {
		userProfile := &pb.UserProfile{
			UserId:      user.userId,
			Username:    user.username,
			PhoneNumber: user.phoneNumber,
			Address: &pb.Address{
				AddressName: user.address.addressName.String,
				SubDistrict: user.address.subDistrict.String,
				District:    user.address.district.String,
				Province:    user.address.province.String,
				PostalCode:  user.address.postalCode.String,
			},
		}
		userProfiles = append(userProfiles, userProfile)
	}

	return &pb.ListUserProfileResponse{UserProfiles: userProfiles}, nil
}

func (x *UserService) DeleteUserProfile(ctx context.Context, in *pb.DeleteUserProfileRequest) (*pb.DeleteUserProfileResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	err := x.repository.DeleteUserProfile(ctx, in.Username)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteUserProfileResponse{Success: true}, nil
}

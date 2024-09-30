package internal

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

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
		return nil, status.Error(codes.InvalidArgument, "username, phone number or address must be provided")
	}

	address := &address{
		addressName: sql.NullString{String: in.Address.AddressName},
		subDistrict: sql.NullString{String: in.Address.SubDistrict},
		district:    sql.NullString{String: in.Address.District},
		province:    sql.NullString{String: in.Address.Province},
		postalCode:  sql.NullString{String: in.Address.PostalCode},
	}
	userID, err := x.repository.SaveUserProfile(ctx, in.Username, in.PhoneNumber, address)
	if err != nil {
		slog.Error("save user profile", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to save user to database")
	}

	return &pb.CreateUserProfileResponse{UserId: userID}, nil
}

func (x *UserService) GetUserProfile(ctx context.Context, in *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {

	if in.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id must be provided")
	}

	user, err := x.repository.UserProfile(ctx, in.UserId)
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
	return &pb.GetUserProfileResponse{UserProfile: userProfile}, nil
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

	if in.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id must be provided")
	}

	err := x.repository.DeleteUserProfile(ctx, in.UserId)
	if err != nil {
		slog.Error("delete user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &pb.DeleteUserProfileResponse{Success: true}, nil
}

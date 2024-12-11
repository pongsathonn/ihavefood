package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

type mockStorage struct{}

func (m mockStorage) Users(ctx context.Context) ([]*dbUserCredentials, error) {
	// Returning a dummy list of dbUserCredentials
	return []*dbUserCredentials{
		{
			UserID:       "2",
			Username:     "user2",
			Email:        "user2@example.com",
			PasswordHash: "",
			Role:         Roles_USER,
			PhoneNumber:  "0987654322",
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		},
		{
			UserID:       "3",
			Username:     "user3",
			Email:        "user3@example.com",
			PasswordHash: "",
			Role:         Roles_USER,
			PhoneNumber:  "0987654232",
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		},
	}, nil
}

func (m mockStorage) User(ctx context.Context, userID string) (*dbUserCredentials, error) {

	if userID == "2" {
		return &dbUserCredentials{
			UserID:       "2",
			Username:     "user2",
			Email:        "user2@example.com",
			PasswordHash: "",
			Role:         Roles_ADMIN,
			PhoneNumber:  "2234567890",
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m mockStorage) UserByUsername(ctx context.Context, username string) (*dbUserCredentials, error) {
	if username == "user2" {
		return &dbUserCredentials{
			UserID:       "2",
			Username:     "user2",
			Email:        "user2@example.com",
			PasswordHash: "",
			Role:         Roles_USER,
			PhoneNumber:  "2234567890",
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}, nil
	}
	return nil, errors.New("user not found")
}

func (m mockStorage) Create(ctx context.Context, newUser *NewUserCredentials) (string, error) {
	userID := "2"
	return userID, nil
}

func (m mockStorage) UpdateRole(ctx context.Context, userID string, newRole dbRoles) (string, error) {
	if userID != "2" {
		return "", errors.New("user not found")
	}
	return userID, nil
}

func (m mockStorage) Delete(ctx context.Context, userID string) error {
	// Simulating successful deletion of a user
	if userID == "2" {
		return nil
	}
	return fmt.Errorf("user not found")
}

func (m mockStorage) ValidateLogin(ctx context.Context, username, password string) (bool, error) {
	// Simulating a valid login check
	if username == "user2" && password == "password" {
		return true, nil
	}
	return false, nil
}

func (m mockStorage) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	// Simulating checking if a username exists
	if username == "user2" {
		return true, nil
	}
	return false, nil
}

type mockProfileClient struct{}

func (m mockProfileClient) ListProfiles(ctx context.Context, in *pb.ListProfilesRequest,
	opts ...grpc.CallOption) (*pb.ListProfilesResponse, error) {

	return nil, nil
}

func (m mockProfileClient) GetProfile(ctx context.Context, in *pb.GetProfileRequest,
	opts ...grpc.CallOption) (*pb.Profile, error) {

	return nil, nil
}

func (m mockProfileClient) CreateProfile(ctx context.Context, in *pb.CreateProfileRequest,
	opts ...grpc.CallOption) (*pb.Profile, error) {

	return &pb.Profile{
		UserId:   in.UserId,
		Username: in.Username,
		Bio:      "",
		Social: &pb.Social{
			Facebook:  "",
			Instagram: "",
			Line:      "",
		},
		Addresses: []*pb.Address{
			&pb.Address{
				AddressName: "",
				SubDistrict: "",
				District:    "",
				Province:    "",
				PostalCode:  "",
			},
		},
		CreateTime: timestamppb.Now(),
		UpdateTime: timestamppb.Now(),
	}, nil
}

func (m mockProfileClient) CreateAddress(ctx context.Context, in *pb.CreateAddressRequest,
	opts ...grpc.CallOption) (*pb.Profile, error) {

	return nil, nil
}

func (m mockProfileClient) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest,
	opts ...grpc.CallOption) (*pb.Profile, error) {

	return nil, nil
}

func (m mockProfileClient) DeleteProfile(ctx context.Context, in *pb.DeleteProfileRequest,
	opts ...grpc.CallOption) (*emptypb.Empty, error) {

	return nil, nil
}

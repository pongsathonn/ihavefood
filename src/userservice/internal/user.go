package internal

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

// UserService manages user profile.
type UserService struct {
	pb.UnimplementedUserServiceServer

	rabbitmq RabbitMQ
	store    UserStorage
}

func NewUserService(rabbitmq RabbitMQ, repo UserStorage) *UserService {
	return &UserService{
		rabbitmq: rabbitmq,
		store:    repo,
	}
}

func (x *UserService) ListProfile(ctx context.Context, in *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {

	// TODO validate input

	results, err := x.store.Profiles(ctx)
	if err != nil {
		return nil, err
	}

	var profiles []*pb.Profile
	for _, profile := range results {
		profiles = append(profiles, dbToProto(profile))
	}

	return &pb.ListProfilesResponse{Profiles: profiles}, nil
}

func (x *UserService) GetProfile(ctx context.Context, in *pb.GetProfileRequest) (*pb.Profile, error) {

	//TODO validate

	profile, err := x.store.Profile(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return dbToProto(profile), nil
}

func (x *UserService) CreateProfile(ctx context.Context, in *pb.CreateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	newProfile := &dbProfile{
		UserID:   in.UserId,
		Username: in.Username,
	}

	userID, err := x.store.Create(ctx, newProfile)
	if err != nil {
		slog.Error("failed to create user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to create user profile")
	}

	profile, err := x.store.Profile(ctx, userID)
	if err != nil {
		slog.Error("failed to retrive user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive user profile")
	}

	return dbToProto(profile), nil
}

func (x *UserService) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	update := &dbProfile{
		Username: in.NewUsername,
		Picture:  in.NewPicture,
		Bio:      in.NewBio,
		Social: &dbSocial{
			Facebook:   in.NewSocial.Facebook,
			Instragram: in.NewSocial.Instagram,
			Line:       in.NewSocial.Line,
		},
		Address: &dbAddress{
			AddressName: sql.NullString{String: in.NewAddress.AddressName},
			SubDistrict: sql.NullString{String: in.NewAddress.SubDistrict},
			District:    sql.NullString{String: in.NewAddress.District},
			Province:    sql.NullString{String: in.NewAddress.Province},
			PostalCode:  sql.NullString{String: in.NewAddress.PostalCode},
		},
	}

	userID, err := x.store.Update(ctx, in.UserId, update)
	if err != nil {
		slog.Error("failed to update profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update profile")
	}

	profile, err := x.store.Profile(ctx, userID)
	if err != nil {
		slog.Error("failed to retrive profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive profile")
	}

	return dbToProto(profile), nil

}

func (x *UserService) DeleteProfile(ctx context.Context, in *pb.DeleteProfileRequest) (*emptypb.Empty, error) {

	//TODO validate intput

	err := x.store.Delete(ctx, in.UserId)
	if err != nil {
		slog.Error("delete user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

func dbToProto(user *dbProfile) *pb.Profile {
	return &pb.Profile{
		UserId:   user.UserID,
		Username: user.Username,
		Picture:  user.Picture,
		Bio:      user.Bio,
		Social: &pb.Social{
			Facebook:  user.Social.Facebook,
			Instagram: user.Social.Instragram,
			Line:      user.Social.Line,
		},
		Address: &pb.Address{
			AddressName: user.Address.AddressName.String,
			SubDistrict: user.Address.SubDistrict.String,
			District:    user.Address.District.String,
			Province:    user.Address.Province.String,
			PostalCode:  user.Address.PostalCode.String,
		},
		CreateTime: user.CreateTime.Unix(),
	}
}

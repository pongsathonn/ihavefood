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
	for _, result := range results {
		profile := &pb.Profile{
			UserId:   result.UserID,
			Username: result.Username,
			Picture:  result.Picture,
			Bio:      result.Bio,
			Social: &pb.Social{
				Facebook:  result.Social.Facebook,
				Instagram: result.Social.Instragram,
				Line:      result.Social.Line,
			},
			Address: &pb.Address{
				AddressName: result.Address.AddressName.String,
				SubDistrict: result.Address.SubDistrict.String,
				District:    result.Address.District.String,
				Province:    result.Address.Province.String,
				PostalCode:  result.Address.PostalCode.String,
			},
			CreateTime: result.CreateTime.Unix(),
		}
		profiles = append(profiles, profile)
	}

	return &pb.ListProfilesResponse{Profiles: profiles}, nil
}

func (x *UserService) GetProfile(ctx context.Context, in *pb.GetProfileRequest) (*pb.Profile, error) {

	if in.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id must be provided")
	}

	result, err := x.store.Profile(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.Profile{
		UserId:   result.UserID,
		Username: result.Username,
		Picture:  result.Picture,
		Bio:      result.Bio,
		Social: &pb.Social{
			Facebook:  result.Social.Facebook,
			Instagram: result.Social.Instragram,
			Line:      result.Social.Line,
		},
		Address: &pb.Address{
			AddressName: result.Address.AddressName.String,
			SubDistrict: result.Address.SubDistrict.String,
			District:    result.Address.District.String,
			Province:    result.Address.Province.String,
			PostalCode:  result.Address.PostalCode.String,
		},
		CreateTime: result.CreateTime.Unix(),
	}, nil
}

func (x *UserService) CreateProfile(ctx context.Context, in *pb.CreateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	newProfile := &dbProfile{
		UserID:   in.UserId,
		Username: in.Username,
	}

	res, err := x.store.Create(ctx, newProfile)
	if err != nil {
		return nil, err
	}

	return &pb.Profile{
		UserId:   res.UserID,
		Username: res.Username,
	}, nil

}

func (x *UserService) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest) (*pb.Profile, error) {

	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "userID must be provided")
	}

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

	profile, err := x.store.Update(ctx, in.UserId, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update profile %v", err)
	}

	return &profile, nil

}

func (x *UserService) DeleteProfile(ctx context.Context, in *pb.DeleteProfileRequest) (*emptypb.Empty, error) {

	if in.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user id must be provided")
	}

	err := x.store.Delete(ctx, in.UserId)
	if err != nil {
		slog.Error("delete user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

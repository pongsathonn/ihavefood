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
		Username: in.Username,
		Picture:  in.Picture,
		Bio:      in.Bio,
		Social: &dbSocial{
			Facebook:   in.Social.Facebook,
			Instragram: in.Social.Instagram,
			Line:       in.Social.Line,
		},
		Address: &dbAddress{
			AddressName: sql.NullString{String: in.Address.AddressName},
			SubDistrict: sql.NullString{String: in.Address.SubDistrict},
			District:    sql.NullString{String: in.Address.District},
			Province:    sql.NullString{String: in.Address.Province},
			PostalCode:  sql.NullString{String: in.Address.PostalCode},
		},
	}

	res, err := x.store.Create(ctx, newProfile)
	if err != nil {
		slog.Error("create user profile", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to insert user profile to database")
	}

	return &pb.Profile{
		UserId:   res.UserID,
		Username: res.Username,
		Picture:  res.Picture,
		Bio:      res.Bio,
		Social: &pb.Social{
			Facebook:  res.Social.Facebook,
			Instagram: res.Social.Instragram,
			Line:      res.Social.Line,
		},
		Address: &pb.Address{
			AddressName: res.Address.AddressName.String,
		},
		CreateTime: res.CreateTime.Unix(),
	}, nil
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

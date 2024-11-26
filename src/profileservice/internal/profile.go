package internal

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/profileservice/genproto"
)

// ProfileService manages user profile.
type ProfileService struct {
	pb.UnimplementedProfileServiceServer

	rabbitmq RabbitMQ
	store    ProfileStorage
}

func NewProfileService(rabbitmq RabbitMQ, repo ProfileStorage) *ProfileService {
	return &ProfileService{
		rabbitmq: rabbitmq,
		store:    repo,
	}
}

func (x *ProfileService) ListProfile(ctx context.Context, in *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {

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

func (x *ProfileService) GetProfile(ctx context.Context, in *pb.GetProfileRequest) (*pb.Profile, error) {

	//TODO validate

	profile, err := x.store.Profile(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return dbToProto(profile), nil
}

func (x *ProfileService) CreateProfile(ctx context.Context, in *pb.CreateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	userID, err := x.store.Create(ctx, &newProfile{
		UserID:   in.UserId,
		Username: in.Username,
	})
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

func (x *ProfileService) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	update := &dbProfile{
		Username: in.NewUsername,
		Bio:      sql.NullString{String: in.NewBio},
		Social: &dbSocial{
			Facebook:   sql.NullString{String: in.NewSocial.Facebook},
			Instragram: sql.NullString{String: in.NewSocial.Instagram},
			Line:       sql.NullString{String: in.NewSocial.Line},
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

func (x *ProfileService) DeleteProfile(ctx context.Context, in *pb.DeleteProfileRequest) (*emptypb.Empty, error) {

	//TODO validate intput

	err := x.store.Delete(ctx, in.UserId)
	if err != nil {
		slog.Error("delete user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

// TODO
func protoToDb(profile *pb.Profile) *dbProfile {

	return &dbProfile{
		//UserID: "",
		Username: profile.Username,
		Bio:      sql.NullString{String: profile.Bio},
		Social: &dbSocial{
			Facebook:   sql.NullString{String: profile.Social.Facebook},
			Instragram: sql.NullString{String: profile.Social.Instagram},
			Line:       sql.NullString{String: profile.Social.Line},
		},
		Address: &dbAddress{
			AddressName: sql.NullString{String: profile.Address.AddressName},
			SubDistrict: sql.NullString{String: profile.Address.SubDistrict},
			District:    sql.NullString{String: profile.Address.District},
			Province:    sql.NullString{String: profile.Address.Province},
			PostalCode:  sql.NullString{String: profile.Address.PostalCode},
		},
		// CreateTime: nil,
	}

}

func dbToProto(profile *dbProfile) *pb.Profile {

	return &pb.Profile{
		UserId:   profile.UserID,
		Username: profile.Username,
		Bio:      profile.Bio.String,
		Social: &pb.Social{
			Facebook:  profile.Social.Facebook.String,
			Instagram: profile.Social.Instragram.String,
			Line:      profile.Social.Line.String,
		},
		Address: &pb.Address{
			AddressName: profile.Address.AddressName.String,
			SubDistrict: profile.Address.SubDistrict.String,
			District:    profile.Address.District.String,
			Province:    profile.Address.Province.String,
			PostalCode:  profile.Address.PostalCode.String,
		},
		CreateTime: profile.CreateTime.Unix(),
	}
}

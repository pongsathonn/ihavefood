package internal

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/profileservice/genproto"
)

// ProfileService manages user profile.
type ProfileService struct {
	pb.UnimplementedProfileServiceServer

	rabbitmq *rabbitMQ
	store    *profileStorage
}

func NewProfileService(rabbitmq *rabbitMQ, store *profileStorage) *ProfileService {
	return &ProfileService{
		rabbitmq: rabbitmq,
		store:    store,
	}
}

func (x *ProfileService) ListProfile(ctx context.Context, in *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {

	// TODO validate input

	results, err := x.store.profiles(ctx)
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

	profile, err := x.store.profile(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return dbToProto(profile), nil
}

func (x *ProfileService) CreateProfile(ctx context.Context, in *pb.CreateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	userID, err := x.store.create(ctx, &newProfile{
		UserID:   in.UserId,
		Username: in.Username,
	})
	if err != nil {
		slog.Error("failed to create user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to create user profile")
	}

	profile, err := x.store.profile(ctx, userID)
	if err != nil {
		slog.Error("failed to retrive user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive user profile")
	}

	return dbToProto(profile), nil
}

func (x *ProfileService) UpdateAddress(ctx context.Context, in *pb.UpdateAddressRequest) (*pb.Profile, error) {

	// TODO validate input

	userID, err := x.store.updateAddress(ctx, in.UserId, &dbAddress{
		AddressName: sql.NullString{String: in.NewAddress.AddressName},
		SubDistrict: sql.NullString{String: in.NewAddress.SubDistrict},
		District:    sql.NullString{String: in.NewAddress.District},
		Province:    sql.NullString{String: in.NewAddress.Province},
		PostalCode:  sql.NullString{String: in.NewAddress.PostalCode},
	})
	if err != nil {
		slog.Error("failed to update profile address", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update adress")
	}

	profile, err := x.store.profile(ctx, userID)
	if err != nil {
		slog.Error("failed to retrive profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive profile")
	}

	return dbToProto(profile), nil
}

func (x *ProfileService) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest) (*pb.Profile, error) {

	// TODO validate input

	update := &dbProfile{
		Username: in.NewUsername,
		Bio:      sql.NullString{String: in.NewBio},
		Social: &dbSocial{
			Facebook:  sql.NullString{String: in.NewSocial.Facebook},
			Instagram: sql.NullString{String: in.NewSocial.Instagram},
			Line:      sql.NullString{String: in.NewSocial.Line},
		},
	}

	userID, err := x.store.update(ctx, in.UserId, update)
	if err != nil {
		slog.Error("failed to update profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update profile")
	}

	profile, err := x.store.profile(ctx, userID)
	if err != nil {
		slog.Error("failed to retrive profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive profile")
	}

	return dbToProto(profile), nil

}

func (x *ProfileService) DeleteProfile(ctx context.Context, in *pb.DeleteProfileRequest) (*emptypb.Empty, error) {

	//TODO validate intput

	err := x.store.remove(ctx, in.UserId)
	if err != nil {
		slog.Error("delete user profile", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

// TODO
func protoToDb(profile *pb.Profile) *dbProfile {

	var addresses []*dbAddress
	for _, a := range profile.Addresses {
		addresses = append(addresses, &dbAddress{
			AddressName: sql.NullString{String: a.AddressName},
			SubDistrict: sql.NullString{String: a.SubDistrict},
			District:    sql.NullString{String: a.District},
			Province:    sql.NullString{String: a.Province},
			PostalCode:  sql.NullString{String: a.PostalCode},
		})
	}

	return &dbProfile{
		//UserID: "",
		Username: profile.Username,
		Bio:      sql.NullString{String: profile.Bio},
		Social: &dbSocial{
			Facebook:  sql.NullString{String: profile.Social.Facebook},
			Instagram: sql.NullString{String: profile.Social.Instagram},
			Line:      sql.NullString{String: profile.Social.Line},
		},
		Addresses: addresses,
		// CreateTime: nil,
	}

}

func dbToProto(profile *dbProfile) *pb.Profile {

	var addresses []*pb.Address
	for _, a := range profile.Addresses {
		addresses = append(addresses, &pb.Address{
			AddressName: a.AddressName.String,
			SubDistrict: a.SubDistrict.String,
			District:    a.District.String,
			Province:    a.Province.String,
			PostalCode:  a.PostalCode.String,
		})
	}

	return &pb.Profile{
		UserId:   profile.UserID,
		Username: profile.Username,
		Bio:      profile.Bio.String,
		Social: &pb.Social{
			Facebook:  profile.Social.Facebook.String,
			Instagram: profile.Social.Instagram.String,
			Line:      profile.Social.Line.String,
		},
		Addresses:  addresses,
		CreateTime: timestamppb.New(profile.CreateTime),
	}
}

package internal

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
)

var (
	errNoRestaurantName = status.Errorf(codes.InvalidArgument, "restaurant name must be provided")
	errNoRestaurantId   = status.Errorf(codes.InvalidArgument, "restaurant id must be provided")
	errNoMenus          = status.Errorf(codes.InvalidArgument, "menu must be at least one")
	errNoFoodName       = status.Errorf(codes.InvalidArgument, "food name must be provided")
	errUnknownTypeMenu  = status.Errorf(codes.InvalidArgument, "menu status cannot be UNKNOWN")
)

type RestaurantService struct {
	pb.UnimplementedRestaurantServiceServer

	repository RestaurantRepository
	rabbitmq   RabbitMQ
}

func NewRestaurantService(repository RestaurantRepository, rabbitmq RabbitMQ) *RestaurantService {
	return &RestaurantService{
		repository: repository,
		rabbitmq:   rabbitmq,
	}
}

func (x *RestaurantService) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantId == "" {
		return nil, errNoRestaurantId
	}

	restaurant, err := x.repository.Restaurant(ctx, in.RestaurantId)
	if err != nil {
		log.Printf("Failed to retrive restaurant: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurant")
	}

	return &pb.GetRestaurantResponse{Restaurant: restaurant}, nil
}

func (x *RestaurantService) ListRestaurant(ctx context.Context, empty *pb.Empty) (*pb.ListRestaurantResponse, error) {

	restaurants, err := x.repository.Restaurants(ctx)
	if err != nil {
		log.Printf("Failed to retrive restaurants: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurants")
	}

	resp := &pb.ListRestaurantResponse{Restaurants: restaurants}

	return resp, nil
}

func (x *RestaurantService) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

	if in.RestaurantName == "" {
		return nil, errNoRestaurantName
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	id, err := x.repository.SaveRestaurant(ctx, in.RestaurantName, in.Menus, in.Address)
	if err != nil {
		log.Printf("Save restaurant failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to save restaurant")
	}

	return &pb.RegisterRestaurantResponse{RestaurantId: id}, nil
}

func (x *RestaurantService) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.AddMenuResponse, error) {

	if in.RestaurantName == "" {
		return nil, errNoRestaurantName
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	if err := x.repository.UpdateMenu(ctx, in.RestaurantName, in.Menus); err != nil {
		log.Printf("Update menu failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to add menu")
	}

	return &pb.AddMenuResponse{Success: true}, nil

}

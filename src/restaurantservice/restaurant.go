package main

import (
	"context"
	"fmt"

	"github.com/pongsathonn/ihavefood/src/restaurantservice/rabbitmq"
	"github.com/pongsathonn/ihavefood/src/restaurantservice/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
)

type restaurant struct {
	pb.UnimplementedRestaurantServiceServer

	rp repository.RestaurantRepo
	mb rabbitmq.RabbitmqClient
}

func NewRestaurant(rp repository.RestaurantRepo, mb rabbitmq.RabbitmqClient) *restaurant {
	return &restaurant{
		rp: rp,
		mb: mb,
	}
}

func (x *restaurant) CheckAvailableMenu(ctx context.Context, in *pb.CheckAvailableMenuRequest) (*pb.CheckAvailableMenuResponse, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or menus must be provided")
	}

	avaliable, err := x.rp.IsAvailableMenu(context.TODO(), in.RestaurantName, in.Menus)
	if err != nil {
		err = status.Errorf(codes.Internal, err.Error())
		return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_UNKNOWN}, err
	}

	if !avaliable {
		return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_UNVAILABLE}, nil
	}

	return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_AVAILABLE}, nil
}

func (x *restaurant) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name must be provided")
	}

	restaurant, err := x.rp.Restaurant(context.TODO(), in.RestaurantName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurant: %v", err)
	}

	return &pb.GetRestaurantResponse{Restaurant: restaurant}, nil
}

func (x *restaurant) ListRestaurant(context.Context, *pb.Empty) (*pb.ListRestaurantResponse, error) {

	restaurants, err := x.rp.Restaurants(context.TODO())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurants: %v", err)
	}

	resp := &pb.ListRestaurantResponse{Restaurants: restaurants}

	return resp, nil
}

func (x *restaurant) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or munus must be provided")
	}

	for _, m := range in.Menus {
		sss := m.Available.String()
		if _, ok := pb.AvailStatus_value[sss]; !ok {
			return nil, fmt.Errorf("menus status invalid")
		}
	}

	id, err := x.rp.SaveRestaurant(context.TODO(), in.RestaurantName, in.Menus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save restaurant: %v", err)
	}

	return &pb.RegisterRestaurantResponse{RestaurantId: id}, nil
}

// TODO
func (x *restaurant) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.Empty, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or munus must be provided")
	}

	if err := x.rp.UpdateMenu(context.TODO(), in.RestaurantName, in.Menus); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update menu: %v", err)
	}

	return &pb.Empty{}, nil

}

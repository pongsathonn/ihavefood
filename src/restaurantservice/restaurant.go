package main

import (
	"context"
	"errors"

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

func (s *restaurant) CheckAvailableMenu(ctx context.Context, in *pb.CheckAvailableMenuRequest) (*pb.CheckAvailableMenuResponse, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "input shouldn't be empty")
	}

	avaliable, err := s.rp.IsAvailableMenu(context.TODO(), in.RestaurantName, in.Menus)
	if err != nil {
		err = status.Errorf(codes.Internal, err.Error())
		return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_UNKNOWN}, err
	}

	if !avaliable {
		return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_UNVAILABLE}, nil
	}

	return &pb.CheckAvailableMenuResponse{Available: pb.AvailStatus_AVAILABLE}, nil
}

func (s *restaurant) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name is empty")
	}

	restaurant, err := s.rp.Restaurant(context.TODO(), in.RestaurantName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error query menus", err)
	}

	return &pb.GetRestaurantResponse{Restaurant: restaurant}, nil
}

func (s *restaurant) ListRestaurant(context.Context, *pb.Empty) (*pb.ListRestaurantResponse, error) {

	restaurants, err := s.rp.Restaurants(context.TODO())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error query restaurants", err)
	}

	resp := &pb.ListRestaurantResponse{Restaurants: restaurants}

	return resp, nil
}

func (s *restaurant) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or menus is empty")
	}

	for _, m := range in.Menus {
		x := m.Available.String()
		if _, ok := pb.AvailStatus_value[x]; !ok {
			return nil, errors.New("menu status invalid")
		}
	}

	id, err := s.rp.SaveRestaurant(context.TODO(), in.RestaurantName, in.Menus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.RegisterRestaurantResponse{RestaurantId: id}, nil
}

// TODO
func (s *restaurant) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.Empty, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or menus is empty")
	}

	if err := s.rp.UpdateMenu(context.TODO(), in.RestaurantName, in.Menus); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.Empty{}, nil

}

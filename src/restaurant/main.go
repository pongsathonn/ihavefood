package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pongsathonn/food-delivery/src/restaurant/pubsub"
	"github.com/pongsathonn/food-delivery/src/restaurant/repository"

	pb "github.com/pongsathonn/food-delivery/src/restaurant/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type restaurantService struct {
	pb.UnimplementedRestaurantServiceServer

	rp repository.RestaurantRepo
	mb pubsub.MessageBroker
}

func NewRestaurantService(rp repository.RestaurantRepo, mb pubsub.MessageBroker) *restaurantService {
	return &restaurantService{
		rp: rp,
		mb: mb,
	}
}

func (s *restaurantService) CheckAvailableMenu(ctx context.Context, in *pb.CheckAvailableMenuRequest) (*pb.CheckAvailableMenuResponse, error) {

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

func (s *restaurantService) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name is empty")
	}

	restaurant, err := s.rp.Restaurant(context.TODO(), in.RestaurantName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error query menus", err)
	}

	return &pb.GetRestaurantResponse{Restaurant: restaurant}, nil
}

func (s *restaurantService) ListRestaurant(context.Context, *pb.Empty) (*pb.ListRestaurantResponse, error) {

	restaurants, err := s.rp.Restaurants(context.TODO())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error query restaurants", err)
	}

	resp := &pb.ListRestaurantResponse{Restaurants: restaurants}

	return resp, nil
}

func (s *restaurantService) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

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
func (s *restaurantService) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.Empty, error) {

	if in.RestaurantName == "" || len(in.Menus) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "restaurant name or menus is empty")
	}

	if err := s.rp.SaveMenu(context.TODO(), in.RestaurantName, in.Menus); err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.Empty{}, nil

}

func initMongoDB() *mongo.Client {
	conn, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("RESTAURANT_DB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	coll := conn.Database("restaurant_database", nil).Collection("restaurantCollection")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"restaurantName", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Println(err)
	}
	return conn
}

func initRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial(os.Getenv("AMQP_URI"))
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func main() {

	rp := repository.NewRestaurantRepo(initMongoDB())
	mb := pubsub.NewMessageBroker(initRabbitMQ())

	rs := NewRestaurantService(rp, mb)
	s := grpc.NewServer()

	pb.RegisterRestaurantServiceServer(s, rs)

	log.Println("restaurant service is running")

	lis, err := net.Listen("tcp", os.Getenv("RESTAURANT_URI"))
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(s.Serve(lis))

}

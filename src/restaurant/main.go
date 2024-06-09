package main

import (
	"context"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db "github.com/pongsathonn/food-delivery/src/restaurant/data"
	pb "github.com/pongsathonn/food-delivery/src/restaurant/genproto"
	"github.com/pongsathonn/food-delivery/src/restaurant/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

type restaurantService struct {
	pb.UnimplementedRestaurantServiceServer

	db db.RestaurantRepo
	mb pubsub.MessageBroker
}

func NewRestaurantService(db db.RestaurantRepo, mb pubsub.MessageBroker) *restaurantService {
	return &restaurantService{
		db: db,
		mb: mb,
	}
}

func (rs *restaurantService) CheckAvaliableMenu(ctx context.Context, in *pb.CheckAvaliableMenuRequest) (*pb.CheckAvaliableMenuResponse, error) {

	if in.Req.RestaurantName == "" || in.Req.Menus == nil {
		return nil, status.Errorf(codes.InvalidArgument, "input shouldn't be empty")
	}

	avaliable, err := rs.db.IsAvaliableMenu(context.TODO(), in)
	if err != nil {
		return &pb.CheckAvaliableMenuResponse{Avaliable: pb.AvaliStatus_UNKNOWN},
			status.Errorf(codes.Internal, err.Error())
	}

	if !avaliable {
		return &pb.CheckAvaliableMenuResponse{Avaliable: pb.AvaliStatus_UNAVALIABLE}, nil
	}

	return &pb.CheckAvaliableMenuResponse{Avaliable: pb.AvaliStatus_AVALIABLE}, nil
}

func main() {

	amqpConn, err := amqp.Dial(os.Getenv("AMQP_URI"))
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("RESTAURANT_DB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	if err := dbConn.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", os.Getenv("RESTAURANT_URI"))

	db := db.NewRestaurantRepo(dbConn)
	mb := pubsub.NewMessageBroker(amqpConn)

	rtsv := NewRestaurantService(db, mb)
	s := grpc.NewServer()

	pb.RegisterRestaurantServiceServer(s, rtsv)

	log.Println("restaurant service is running")

	log.Fatal(s.Serve(lis))

}

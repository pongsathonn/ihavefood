package data

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/food-delivery/src/restaurant/genproto"
)

type RestaurantRepo interface {
	IsAvaliableMenu(context.Context, *pb.CheckAvaliableMenuRequest) (bool, error)
	ListMenu(context.Context, *pb.ListMenuRequest) (*pb.ListMenuResponse, error)
	ListRestaurant(context.Context, *pb.Empty) (*pb.ListRestaurantResponse, error)
	NewRestaurant(context.Context, *pb.NewRestaurantRequest) (*pb.NewRestaurantResponse, error)
	AddMenu(context.Context, *pb.AddMenuRequest) (*pb.Empty, error)
}

type Restaurant struct {
	RestaurantId   primitive.ObjectID `bson:"_id,omitempty"`
	RestaurantName string             `bson:"restaurantName"`
	Menus          []*pb.Menu         `bson:"menus"`
}

type restaurantRepo struct {
	db *mongo.Client
}

func NewRestaurantRepo(db *mongo.Client) RestaurantRepo {
	return &restaurantRepo{db: db}
}

func (rs *restaurantRepo) IsAvaliableMenu(ctx context.Context, in *pb.CheckAvaliableMenuRequest) (bool, error) {
	coll := rs.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	var menus []pb.Menu

	filter := bson.D{{"restaurantName", in.Req.RestaurantName}}
	if err := coll.FindOne(context.TODO(), filter).Decode(&menus); err != nil {
		return false, err
	}

	// 0 ava, 1 unav , 2 unknown
	for _, menu := range menus {
		if menu.Avaliable.Number() == 1 {
			err := fmt.Errorf("menu %s is unavaliable", menu.FoodName)
			return false, err
		}
	}

	return true, nil
}

func (rs *restaurantRepo) ListMenu(ctx context.Context, in *pb.ListMenuRequest) (*pb.ListMenuResponse, error) {
	log.Println("list menu ja")
	return nil, nil
}

func (rs *restaurantRepo) ListRestaurant(context.Context, *pb.Empty) (*pb.ListRestaurantResponse, error) {
	log.Println("list restaurant ja")
	return nil, nil
}

func (rs *restaurantRepo) NewRestaurant(ctx context.Context, in *pb.NewRestaurantRequest) (*pb.NewRestaurantResponse, error) {

	coll := rs.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	newId := primitive.NewObjectID()

	restau := Restaurant{
		RestaurantId:   newId,
		RestaurantName: in.Req.RestaurantName,
		Menus:          in.Req.Menus,
	}

	res, err := coll.InsertOne(context.TODO(), restau)
	if err != nil {
		return nil, err
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("error assert type")
	}

	return &pb.NewRestaurantResponse{RestaurantId: id.Hex()}, nil
}

func (rs *restaurantRepo) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.Empty, error) {

	coll := rs.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	restauName := in.Req.RestaurantName
	filter := bson.D{{"restaurantName", restauName}}
	_, err := coll.UpdateOne(context.TODO(), filter, in)
	if err != nil {
		return nil, err
	}

	log.Println("new menu has added")

	return nil, nil
}

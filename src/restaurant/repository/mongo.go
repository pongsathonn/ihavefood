package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/food-delivery/src/restaurant/genproto"
)

type RestaurantRepo interface {
	IsAvailableMenu(ctx context.Context, restauName string, menus []*pb.Menu) (bool, error)
	Restaurant(ctx context.Context, restauName string) (*pb.Restaurant, error)
	Restaurants(context.Context) ([]*pb.Restaurant, error)
	SaveRestaurant(ctx context.Context, restauName string, menus []*pb.Menu) (string, error)
	SaveMenu(ctx context.Context, restauName string, menus []*pb.Menu) error
}

type RestaurantEntity struct {
	RestaurantId   primitive.ObjectID `bson:"_id,omitempty"`
	RestaurantName string             `bson:"restaurantName"`
	Menus          []*pb.Menu         `bson:"menus"`
}

// 0 avaiable, 1 unavailable, 2 unknown
type MenuEntity struct {
	FoodName  string         `bson:"foodName,omitempty"`
	Price     int32          `bson:"price,omitempty"`
	Available pb.AvailStatus `bson:"available,omitempty"`
}

type restaurantRepo struct {
	db *mongo.Client
}

func NewRestaurantRepo(db *mongo.Client) RestaurantRepo {
	return &restaurantRepo{db: db}
}

func (rp *restaurantRepo) IsAvailableMenu(ctx context.Context, restauName string, menus []*pb.Menu) (bool, error) {
	coll := rp.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	if len(menus) == 0 {
		err := errors.New("menus is empty")
		return false, err
	}

	filter := bson.D{{"restaurantName", restauName}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return false, err
	}
	defer cursor.Close(context.TODO())

	var resp []*MenuEntity
	if err := cursor.All(context.TODO(), &resp); err != nil {
		return false, err
	}

	if len(menus) == 0 {
		return false, errors.New("no documents ja")
	}

	// 0 avaliable, 1 unavaliable , 2 unknown
	for _, mn := range menus {
		if mn.Available != 0 {
			return false, errors.New("menu not avaliable ")
		}
	}

	return true, nil
}

func (rp *restaurantRepo) Restaurant(ctx context.Context, restauName string) (*pb.Restaurant, error) {

	if restauName == "" {
		return nil, errors.New("empty ja")
	}

	coll := rp.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	filter := bson.M{"restaurantName": restauName}

	var rx RestaurantEntity
	err := coll.FindOne(context.TODO(), filter).Decode(&rx)
	if err != nil {
		return nil, err
	}

	return &pb.Restaurant{
		RestaurantId:   rx.RestaurantId.Hex(),
		RestaurantName: rx.RestaurantName,
		Menus:          rx.Menus,
	}, nil
}

func (rp *restaurantRepo) Restaurants(context.Context) ([]*pb.Restaurant, error) {

	coll := rp.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	filter := bson.D{{}}
	cur, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	var resp []*pb.Restaurant
	for cur.Next(context.TODO()) {
		var rx RestaurantEntity
		if err := cur.Decode(&rx); err != nil {
			return nil, err
		}
		resp = append(resp, &pb.Restaurant{
			RestaurantId:   rx.RestaurantId.Hex(),
			RestaurantName: rx.RestaurantName,
			Menus:          rx.Menus,
		})
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return resp, nil

}

func (rp *restaurantRepo) SaveRestaurant(ctx context.Context, restauName string, menus []*pb.Menu) (string, error) {

	coll := rp.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	restau := RestaurantEntity{
		RestaurantId:   primitive.NewObjectID(),
		RestaurantName: restauName,
		Menus:          menus,
	}

	res, err := coll.InsertOne(context.TODO(), restau)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("restaurant name already exists")
		}
		return "", err
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("error assert type")
	}

	return id.Hex(), nil

}

func (rp *restaurantRepo) SaveMenu(ctx context.Context, restauName string, newMenus []*pb.Menu) error {

	coll := rp.db.Database("restaurant_database", nil).Collection("restaurantCollection")

	coll.Find(context.TODO(), nil)

	filter := bson.D{{"restaurantName", restauName}}
	update := bson.D{{"$push", bson.D{{"menus", bson.D{{"$each", newMenus}}}}}}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

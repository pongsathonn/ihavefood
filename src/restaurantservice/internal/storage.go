package internal

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
)

type RestaurantStorage interface {
	Restaurant(ctx context.Context, restaurantNO string) (*dbRestaurant, error)
	Restaurants(ctx context.Context) ([]*dbRestaurant, error)
	SaveRestaurant(ctx context.Context, restaurantName string, menus []*dbMenu, address *dbAddress) (string, error)
	UpdateMenu(ctx context.Context, restaurantId string, menus []*dbMenu) error
}

type restaurantStorage struct {
	client *mongo.Client
}

func NewRestaurantStorage(client *mongo.Client) RestaurantStorage {
	return &restaurantStorage{client: client}
}

func (s *restaurantStorage) Restaurant(ctx context.Context, restaurantNO string) (*dbRestaurant, error) {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	if restaurantId == "" {
		return nil, errors.New("restaurant id must be provided")
	}

	filter := bson.M{"_id": restaurantId}
	var restaurant dbRestaurant

	err := coll.FindOne(ctx, filter).Decode(&restaurant)
	if err != nil {
		return nil, err
	}

	return &pb.Restaurant{
		No:      restaurant.No.Hex(),
		Name:    restaurant.Name,
		Menus:   restaurant.Menus,
		Address: restaurant.Address,
	}, nil

}

func (s *restaurantStorage) Restaurants(ctx context.Context) ([]*dbRestaurant, error) {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var restaurants []*pb.Restaurant

	for cursor.Next(ctx) {

		var restaurant dbRestaurant
		if err := cursor.Decode(&restaurant); err != nil {
			return nil, err
		}

		restaurants = append(restaurants, &pb.Restaurant{
			No:      restaurant.RestaurantNo.Hex(),
			Name:    restaurant.RestaurantName,
			Menus:   restaurant.Menus,
			Address: restaurant.Address,
		})
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return restaurants, nil

}

// ignore a restaurant number and let Mongo generate _id
func (s *restaurantStorage) SaveRestaurant(ctx context.Context, restaurantName string, menus []*dbMenu, address *dbAddress) (string, error) {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	res, err := coll.InsertOne(ctx, dbRestaurant{
		//No:   primitive.NewObjectID(),
		RestaurantName: restaurantName,
		Menus:          menus,
		Address:        address,
		Status:         pb.Status(pb.Status_OPEN),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("restaurant name already exists")
		}
		return "", err
	}

	restaurantNO, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("number not primitive.ObjectID type")
	}

	return restaurantNO.Hex(), nil

}

func (s *restaurantStorage) UpdateMenu(ctx context.Context, restaurantNO string, menus []*dbMenu) error {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	no, err := primitive.ObjectIDFromHex(restaurantNO)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": no}
	update := bson.M{"$push": bson.M{"menus": bson.M{"$each": newMenus}}}

	res, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("restaurant not found")
	}

	if res.ModifiedCount == 0 {
		return errors.New("update menu failed")
	}

	return nil
}

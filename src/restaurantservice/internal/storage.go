package internal

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RestaurantStorage interface {
	Restaurant(ctx context.Context, restaurantNO string) (*dbRestaurant, error)

	Restaurants(ctx context.Context) ([]*dbRestaurant, error)

	SaveRestaurant(ctx context.Context, newRestaurant *newRestaurant) (string, error)

	UpdateMenu(ctx context.Context, restaurantNO string, newMenus []*dbMenu) (string, error)
}

type restaurantStorage struct {
	client *mongo.Client
}

func NewRestaurantStorage(client *mongo.Client) RestaurantStorage {
	return &restaurantStorage{client: client}
}

func (s *restaurantStorage) Restaurant(ctx context.Context, restaurantNO string) (*dbRestaurant, error) {

	coll := s.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	ID, err := primitive.ObjectIDFromHex(restaurantNO)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": ID}

	var restaurant dbRestaurant
	if err := coll.FindOne(ctx, filter).Decode(&restaurant); err != nil {
		return nil, err
	}

	return &restaurant, nil
}

func (s *restaurantStorage) Restaurants(ctx context.Context) ([]*dbRestaurant, error) {

	coll := s.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var restaurants []*dbRestaurant
	for cursor.Next(ctx) {

		var restaurant dbRestaurant
		if err := cursor.Decode(&restaurant); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &restaurant)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return restaurants, nil

}

// ignore a restaurant number and let Mongo generate _id
func (s *restaurantStorage) SaveRestaurant(ctx context.Context,
	newRestaurant *newRestaurant) (string, error) {

	coll := s.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	res, err := coll.InsertOne(ctx, dbRestaurant{
		Name:    newRestaurant.RestaurantName,
		Menus:   newRestaurant.Menus,
		Address: newRestaurant.address,
		Status:  dbStatus(Status_CLOSED),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("restaurant name already exists")
		}
		return "", err
	}

	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return insertedID.Hex(), nil

}

func (s *restaurantStorage) UpdateMenu(ctx context.Context, restaurantNO string, newMenus []*dbMenu) (string, error) {

	coll := s.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	ID, err := primitive.ObjectIDFromHex(restaurantNO)
	if err != nil {
		return "", err
	}

	update := bson.M{"$push": bson.M{"menus": bson.M{"$each": newMenus}}}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		return "", errors.New("restaurant not found")
	}

	upsertedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return upsertedID.Hex(), nil
}

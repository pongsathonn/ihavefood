package internal

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
)

type RestaurantRepository interface {
	Restaurant(ctx context.Context, restaurantId string) (*pb.Restaurant, error)
	Restaurants(ctx context.Context) ([]*pb.Restaurant, error)
	SaveRestaurant(ctx context.Context, restaurantName string, menus []*pb.Menu, address *pb.Address) (string, error)
	UpdateMenu(ctx context.Context, restaurantId string, menus []*pb.Menu) error
}

// RestaurantEntity represents a MongoDB document for restaurants.
// The RestaurantId field maps to the MongoDB _id, which is auto-generated
// and used as the restaurant's unique identifier.
type RestaurantEntity struct {
	RestaurantNo   primitive.ObjectID `bson:"_id,omitempty"`
	RestaurantName string
	Menus          []*pb.Menu
	Address        *pb.Address
	Status         pb.Status
}

type restaurantRepository struct {
	client *mongo.Client
}

func NewRestaurantRepository(client *mongo.Client) RestaurantRepository {
	return &restaurantRepository{client: client}
}

func (r *restaurantRepository) Restaurant(ctx context.Context, restaurantId string) (*pb.Restaurant, error) {

	var entity RestaurantEntity
	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	if restaurantId == "" {
		return nil, errors.New("restaurant id must be provided")
	}

	filter := bson.M{"_id": restaurantId}
	err := coll.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		return nil, err
	}

	restaurant := &pb.Restaurant{
		No:      entity.RestaurantNo.Hex(),
		Name:    entity.RestaurantName,
		Menus:   entity.Menus,
		Address: entity.Address,
	}

	return restaurant, nil
}

func (r *restaurantRepository) Restaurants(ctx context.Context) ([]*pb.Restaurant, error) {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var restaurants []*pb.Restaurant
	for cursor.Next(ctx) {
		var entity RestaurantEntity
		if err := cursor.Decode(&entity); err != nil {
			return nil, err
		}
		restaurant := &pb.Restaurant{
			No:      entity.RestaurantNo.Hex(),
			Name:    entity.RestaurantName,
			Menus:   entity.Menus,
			Address: entity.Address,
		}
		restaurants = append(restaurants, restaurant)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return restaurants, nil

}

func (r *restaurantRepository) SaveRestaurant(
	ctx context.Context,
	restaurantName string,
	menus []*pb.Menu,
	address *pb.Address,
) (string, error) {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	// ignore a restaurant number and let Mongo generate _id
	restaurant := RestaurantEntity{
		//No:   primitive.NewObjectID(),
		RestaurantName: restaurantName,
		Menus:          menus,
		Address:        address,
		Status:         pb.Status(pb.Status_OPEN),
	}

	res, err := coll.InsertOne(ctx, restaurant)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("restaurant name already exists")
		}
		return "", err
	}

	no, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("number not primitive.ObjectID type")
	}

	return no.Hex(), nil

}

func (r *restaurantRepository) UpdateMenu(ctx context.Context, restaurantNo string, newMenus []*pb.Menu) error {

	coll := r.client.Database("restaurant_database", nil).Collection("restaurantCollection")

	no, err := primitive.ObjectIDFromHex(restaurantNo)
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

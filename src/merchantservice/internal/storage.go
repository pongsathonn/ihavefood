package internal

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type merchantStorage struct {
	client *mongo.Client
}

func NewMerchantStorage(client *mongo.Client) MerchantStorage {
	return &merchantStorage{client: client}
}

func (s *merchantStorage) GetMerchant(ctx context.Context, merchantID string) (*dbMerchant, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	var merchant dbMerchant
	if err := coll.FindOne(ctx, bson.M{"_id": merchantID}).Decode(&merchant); err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (s *merchantStorage) ListMerchants(ctx context.Context) ([]*dbMerchant, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var merchants []*dbMerchant
	for cursor.Next(ctx) {

		var merchant dbMerchant
		if err := cursor.Decode(&merchant); err != nil {
			return nil, err
		}
		merchants = append(merchants, &merchant)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return merchants, nil

}

func (s *merchantStorage) SaveMerchant(ctx context.Context, merchantID string, merchantName string) (*dbMerchant, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	res, err := coll.InsertOne(ctx, dbMerchant{
		ID:   merchantID,
		Name: merchantName,
	})
	if err != nil {
		return nil, err
	}

	var merchant *dbMerchant
	if err := coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&merchant); err != nil {
		return nil, err
	}

	return merchant, nil
}

func (s *merchantStorage) CreateMenu(ctx context.Context, merchantID string, menu []*dbMenuItem) ([]*dbMenuItem, error) {
	return nil, errors.New("TODO: CreateMenu not implement")
}

// UpdateMenuItem updates a specific menu item in a merchant's menu
func (s *merchantStorage) UpdateMenuItem(ctx context.Context, merchantID string, updateMenu *dbMenuItem) (*dbMenuItem, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	set := bson.M{}
	if updateMenu.FoodName != "" {
		set["menu.$.food_name"] = updateMenu.FoodName
	}
	if updateMenu.Price != 0 {
		set["menu.$.price"] = updateMenu.Price
	}
	if updateMenu.Description != "" {
		set["menu.$.description"] = updateMenu.Description
	}
	set["menu.$.is_available"] = updateMenu.IsAvailable

	if len(set) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	filter := bson.M{
		"_id":          merchantID,
		"menu.item_id": updateMenu.ItemID,
	}

	var updatedMerchant dbMerchant

	update := bson.M{"$set": set}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(updatedMerchant); err != nil {
		return nil, fmt.Errorf("failed to update menu item: %v", err)
	}

	for _, updatedMenu := range updatedMerchant.Menu {
		if updateMenu.ItemID == updatedMenu.ItemID {
			return updateMenu, nil
		}
	}

	return nil, nil
}

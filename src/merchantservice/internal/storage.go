package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type merchantStorage struct {
	client *mongo.Client
}

func NewMerchantStorage(client *mongo.Client) MerchantStorage {
	return &merchantStorage{client: client}
}

func (s *merchantStorage) MerchantExistsByName(ctx context.Context, name string) (bool, error) {
	coll := s.client.Database("db").Collection("merchants")

	filter := bson.M{"name": name}
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *merchantStorage) GetMerchant(ctx context.Context, merchantID string) (*DbMerchant, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	var merchant DbMerchant
	if err := coll.FindOne(ctx, bson.M{"_id": merchantID}).Decode(&merchant); err != nil {
		return nil, err
	}
	return &merchant, nil
}

func (s *merchantStorage) ListMerchants(ctx context.Context) ([]*DbMerchant, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var merchants []*DbMerchant
	for cursor.Next(ctx) {

		var merchant DbMerchant
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

func (s *merchantStorage) CreateMerchant(ctx context.Context, newMerchant *NewMerchant) (string, error) {

	var merchant DbMerchant
	merchant.ID = uuid.New().String()
	merchant.Name = newMerchant.Name
	for _, item := range newMerchant.Menu {
		merchant.Menu = append(merchant.Menu, &DbMenuItem{
			ItemID:    uuid.New().String(),
			FoodName:  item.FoodName,
			Price:     item.Price,
			ImageInfo: item.ImageInfo,
		})
	}
	if newMerchant.Address != nil {
		merchant.Address = &DbAddress{
			AddressID:   uuid.New().String(),
			AddressName: newMerchant.Address.AddressName,
			SubDistrict: newMerchant.Address.SubDistrict,
			District:    newMerchant.Address.District,
			Province:    newMerchant.Address.Province,
			PostalCode:  newMerchant.Address.PostalCode,
		}
	}
	merchant.ImageInfo = newMerchant.ImageInfo
	merchant.Phone = newMerchant.Phone
	merchant.Email = newMerchant.Email
	merchant.Status = newMerchant.Status

	coll := s.client.Database("db").Collection("merchants")
	_, err := coll.InsertOne(ctx, merchant)
	if err != nil {
		return "", err
	}
	return merchant.ID, nil
}

func (s *merchantStorage) CreateMenu(ctx context.Context, merchantID string, menu []*DbMenuItem) ([]*DbMenuItem, error) {
	return nil, errors.New("TODO: CreateMenu not implement")
}

// UpdateMenuItem updates a specific menu item in a merchant's menu
func (s *merchantStorage) UpdateMenuItem(ctx context.Context, merchantID string, updateMenu *DbMenuItem) (*DbMenuItem, error) {

	coll := s.client.Database("db", nil).Collection("merchants")

	set := bson.M{}
	if updateMenu.FoodName != "" {
		set["menu.$.food_name"] = updateMenu.FoodName
	}
	if updateMenu.Price != 0 {
		set["menu.$.price"] = updateMenu.Price
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	filter := bson.M{
		"_id":          merchantID,
		"menu.item_id": updateMenu.ItemID,
	}

	var updatedMerchant DbMerchant

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

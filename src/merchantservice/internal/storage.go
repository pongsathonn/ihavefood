package internal

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MerchantStorage interface {
	Merchant(ctx context.Context, merchantNO string) (*dbMerchant, error)

	Merchants(ctx context.Context) ([]*dbMerchant, error)

	SaveMerchant(ctx context.Context, newMerchant *newMerchant) (string, error)

	UpdateMenu(ctx context.Context, merchantNO string, newMenu []*dbMenuItem) (string, error)
}

type merchantStorage struct {
	client *mongo.Client
}

func NewMerchantStorage(client *mongo.Client) MerchantStorage {
	return &merchantStorage{client: client}
}

func (s *merchantStorage) Merchant(ctx context.Context, merchantNO string) (*dbMerchant, error) {

	coll := s.client.Database("merchant_database", nil).Collection("merchantCollection")

	ID, err := primitive.ObjectIDFromHex(merchantNO)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": ID}

	var merchant dbMerchant
	if err := coll.FindOne(ctx, filter).Decode(&merchant); err != nil {
		return nil, err
	}

	return &merchant, nil
}

func (s *merchantStorage) Merchants(ctx context.Context) ([]*dbMerchant, error) {

	coll := s.client.Database("merchant_database", nil).Collection("merchantCollection")

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

// ignore a merchant number and let Mongo generate _id
func (s *merchantStorage) SaveMerchant(ctx context.Context,
	newMerchant *newMerchant) (string, error) {

	coll := s.client.Database("merchant_database", nil).Collection("merchantCollection")

	res, err := coll.InsertOne(ctx, dbMerchant{
		Name:    newMerchant.MerchantName,
		Menu:    newMerchant.Menu,
		Address: newMerchant.Address,
		Status:  dbStoreStatus(StoreStatus_CLOSED),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("merchant name already exists")
		}
		return "", err
	}

	insertedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return insertedID.Hex(), nil

}

func (s *merchantStorage) UpdateMenu(ctx context.Context, merchantNO string, newMenu []*dbMenuItem) (string, error) {

	coll := s.client.Database("merchant_database", nil).Collection("merchantCollection")

	ID, err := primitive.ObjectIDFromHex(merchantNO)
	if err != nil {
		return "", err
	}

	update := bson.M{"$push": bson.M{"menus": bson.M{"$each": newMenu}}}

	res, err := coll.UpdateByID(ctx, ID, update)
	if err != nil {
		return "", err
	}

	if res.ModifiedCount == 0 {
		return "", errors.New("merchant not found")
	}

	upsertedID, ok := res.UpsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("ID not primitive.ObjectID")
	}

	return upsertedID.Hex(), nil
}

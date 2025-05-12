// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type newRestaurant struct {
	RestaurantName string
	Menus          []*dbMenu
	address        *dbAddress
}

type dbRestaurant struct {
	No      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	Menus   []*dbMenu          `bson:"menus"`
	Address *dbAddress         `bson:"address"`
	Status  dbStatus           `bson:"status"`
}

type dbMenu struct {
	FoodName string `bson:"foodName"`
	Price    int32  `bson:"price"`
}

type dbAddress struct {
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

type dbStatus int32

const (
	Status_CLOSED dbStatus = 0
	Status_OPEN   dbStatus = 1
)

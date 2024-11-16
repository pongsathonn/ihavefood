// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type dbRestaurant struct {
	No      primitive.ObjectID `bson:"_id,omitempty"`
	Name    string
	Menus   []*dbMenu
	Address *dbAddress
	Status  dbStatus
}

type dbMenu struct {
	FoodName string
	Price    int32
}

type dbAddress struct {
	AddressName string
	SubDistrict string
	District    string
	Province    string
	PostalCode  string
}

type dbStatus int32

const (
	Status_CLOSED dbStatus = 0
	Status_OPEN   dbStatus = 1
)

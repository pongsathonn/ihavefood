package internal

import (
// "go.mongodb.org/mongo-driver/bson/primitive"
)

// dbUpdateMerchant ?
// type dbNewMerchant struct {
// 	MerchantName string        `json:"merchantName"`
// 	Menu         []*dbMenuItem `json:"menu"`
// 	Address      *dbAddress    `json:"address"`
// 	PhoneNumber  string        `json:"phoneNumber"`
// 	Status       dbStoreStatus `json:"status"`
// }

type dbMerchant struct {
	ID          string        `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Menu        []*dbMenuItem `bson:"menu"`
	Address     *dbAddress    `bson:"address"`
	PhoneNumber string        `bson:"phoneNumber"`
	Status      dbStoreStatus `bson:"status"`
}

type dbMenuItem struct {
	ItemID      string `bson:"item_id"`
	FoodName    string `bson:"foodName"`
	Price       int32  `bson:"price"`
	Description string `bson:"description"`
	IsAvailable bool   `bson:"isAvailable"`
}

type dbAddress struct {
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

type dbStoreStatus int32

const (
	StoreStatus_CLOSED dbStoreStatus = 0
	StoreStatus_OPEN   dbStoreStatus = 1
)

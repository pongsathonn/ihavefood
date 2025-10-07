// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"time"
)

type dbNewCustomer struct {
	CustomerID string
	Username   string
	Email      string
	CreateTime time.Time
}

type dbCustomer struct {
	CustomerID string
	Username   string
	Email      string
	Phone      string
	//Picture    []byte
	Social     dbSocial
	Addresses  []*dbAddress
	CreateTime time.Time
	UpdateTime time.Time
}

type dbSocial struct {
	Facebook  *string
	Instagram *string
	Line      *string
}

type dbAddress struct {
	AddressID   string
	AddressName *string
	SubDistrict *string
	District    *string
	Province    *string
	PostalCode  *string
}

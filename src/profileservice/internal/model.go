// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"database/sql"
	"time"
)

type newProfile struct {
	UserID   string
	Username string
}

// UserID and Username is generated from AuthService to ensure
// both user credentials and profile are sync.
type dbProfile struct {
	UserID   string
	Username string
	//Picture    []byte
	Bio        sql.NullString
	Social     dbSocial
	Addresses  []*dbAddress
	CreateTime time.Time
	UpdateTime time.Time
}

type dbSocial struct {
	Facebook  sql.NullString
	Instagram sql.NullString
	Line      sql.NullString
}

type dbAddress struct {
	AddressName sql.NullString
	SubDistrict sql.NullString
	District    sql.NullString
	Province    sql.NullString
	PostalCode  sql.NullString
}

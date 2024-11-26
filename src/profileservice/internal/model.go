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
//
// TODO: Add 'Picture' field to store the user's profile image once
// a storage solution is decided (e.g., local file system, cloud storage).
type dbProfile struct {
	UserID   string
	Username string
	//Picture    []byte
	Bio        sql.NullString
	Social     *dbSocial
	Address    *dbAddress
	CreateTime time.Time
}

type dbSocial struct {
	Facebook   sql.NullString
	Instragram sql.NullString
	Line       sql.NullString
}

type dbAddress struct {
	AddressName sql.NullString
	SubDistrict sql.NullString
	District    sql.NullString
	Province    sql.NullString
	PostalCode  sql.NullString
}

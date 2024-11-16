// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"database/sql"
	"time"
)

// UserID and Username is generated from AuthService
// to ensure sync both user credentials and user Profile
type dbProfile struct {
	UserID     string
	Username   string
	Picture    []byte
	Bio        string
	Social     *dbSocial
	Address    *dbAddress
	CreateTime time.Time
}

type dbSocial struct {
	Facebook   string
	Instragram string
	Line       string
}

type dbAddress struct {
	AddressName sql.NullString
	SubDistrict sql.NullString
	District    sql.NullString
	Province    sql.NullString
	PostalCode  sql.NullString
}

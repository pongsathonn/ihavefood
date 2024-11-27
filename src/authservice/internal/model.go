// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

// NewUserCredential contains information to create
// both new user credential and admin
type NewUserCredentials struct {
	Username    string
	Email       string
	Password    string
	PhoneNumber string
	Role        dbRoles
}

// dbUserCredentials contains ... TODO .
//
// NOTE: PasswordHash must not be contained
// in response
type dbUserCredentials struct {
	UserID       string
	Username     string
	Email        string
	PasswordHash string
	Role         dbRoles
	PhoneNumber  string
	CreateTime   time.Time
	UpdateTime   time.Time
}

type dbRoles int32

const (
	Roles_VISITOR dbRoles = 0
	Roles_USER    dbRoles = 1
	Roles_ADMIN   dbRoles = 2
)

// JWT signing key
var signingKey []byte

// AuthClaims is custom claims use when
// register new jwt claims.
type AuthClaims struct {
	// ID is used as unique identifier for User
	// or Admin depending on the Role.
	ID   string
	Role pb.Roles `json:"role"`
	jwt.RegisteredClaims
}

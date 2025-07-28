// This file contains the structure need for moving data between
// the app and the database
package internal

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

// FIXME this might not be good for prod, find other solution .
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

// dbNewUserCredential contains information to create
// both new user credential and admin
type dbNewUserCredentials struct {
	Username    string
	Email       string
	HashedPass  string
	PhoneNumber string
	Role        dbRoles
}

type dbUserCredentials struct {
	ID          uuid.UUID
	Username    string
	Email       string
	HashedPass  string
	Role        dbRoles
	PhoneNumber string
	CreateTime  time.Time
	UpdateTime  time.Time
}

type dbRoles int32

const (
	Roles_UNKNOWN  dbRoles = 0
	Roles_CUSTOMER dbRoles = 1
	Roles_MERCHANT dbRoles = 2
	Roles_RIDER    dbRoles = 3

	// For simplicity, admin roles are included in this enum.
	Roles_SUPER_ADMIN dbRoles = 20
	Roles_ADMIN       dbRoles = 21
)

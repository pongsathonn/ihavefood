// This file contains the structure need for moving data between
// the app and the database
package internal

import "time"

// NewUserCredential contains information to create new user
// credential
type NewUserCredentials struct {
	Username    string
	Email       string
	Password    string
	PhoneNumber string
}

type dbUserCredentials struct {
	UserID       string
	Username     string
	Email        string
	PasswordHash string
	Role         dbRoles
	PhoneNumber  string
	CreateTime   time.Time
}

type dbRoles int32

const (
	Roles_VISITOR dbRoles = 0
	Roles_USER    dbRoles = 1
	Roles_ADMIN   dbRoles = 2
)

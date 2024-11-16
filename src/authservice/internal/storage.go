package internal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthStorage interface {

	// Users returns a list of user credentials.
	Users(ctx context.Context) ([]*dbUserCredentials, error)

	// User returns the user credentials.
	User(ctx context.Context, userID string) (*dbUserCredentials, error)

	// UserByUsername finds by username and return the user credentials.
	UserByUsername(ctx context.Context, username string) (*dbUserCredentials, error)

	// Create creates new user credential and return its ID.
	Create(ctx context.Context, newUser *NewUserCredentials) (string, error)

	// UpdateRole updates user role and returns ID.
	UpdateRole(ctx context.Context, userID string, newRole dbRoles) (string, error)

	// Delete deletes the user credential.
	Delete(ctx context.Context, userID string) error

	// ValidateLogin checks if the provided username and password are correct.
	// It returns true if the credentials are valid, or false if they are not.
	// Any non-nil error returned is related to the query execution not the
	// validation process. Callers should check the error before interpreting
	// the boolean result.
	ValidateLogin(ctx context.Context, username, password string) (bool, error)

	// CheckUsernameExists checks username already exists in the database.
	CheckUsernameExists(ctx context.Context, username string) (bool, error)

	// CheckExists checks if a user with the given username, email, or phone number
	// already exists. It returns the name of the first field that already exists with
	// a nil error. If the query operation fails, it returns an empty string and an error.
	// Returns an empty string and nil if no fields are found in used.
	CheckExists(ctx context.Context, username, email, phoneNumber string) (string, error)
}

type authStorage struct {
	db *sql.DB
}

func NewAuthStorage(db *sql.DB) AuthStorage {
	return &authStorage{db: db}
}

func (s *authStorage) Users(ctx context.Context) ([]*dbUserCredentials, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			id,
			username, 
			email,
			role
			phone_number,
			create_time
		FROM 
			user_credentials 
	`)
	if err != nil {
		return nil, err
	}

	var users []*dbUserCredentials
	for rows.Next() {
		var user dbUserCredentials
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.PhoneNumber,
			&user.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (s *authStorage) User(ctx context.Context, userID string) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			id,
			username, 
			email,
			role
			phone_number,
			create_time
		FROM 
			user_credentials
		WHERE
			id=$1
	`,
		userID)

	var user dbUserCredentials
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (s *authStorage) UserByUsername(ctx context.Context, username string) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			id,
			username, 
			email,
			role
			phone_number,
			create_time
		FROM 
			user_credentials
		WHERE
			username=$1
	`,
		username)

	var user dbUserCredentials
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (s *authStorage) Create(ctx context.Context, newUser *NewUserCredentials) (string, error) {

	hashedPass, err := hashPassword(newUser.Password)
	if err != nil {
		return "", err
	}

	createTime := time.Now()

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO user_credentials(
			username,
			email,
			password,
			role,
			phone_number,
			create_time
		)
		VALUES($1, $2, $3, $4,$5)
		RETURNING id
	`,
		newUser.Username,
		newUser.Email,
		string(hashedPass),
		newUser.Role,
		createTime,
	)

	var userID string
	if err := row.Scan(&userID); err != nil {
		return "", err
	}

	return userID, nil
}

func (s *authStorage) UpdateRole(ctx context.Context, userID string, newRole dbRoles) (string, error) {

	query := `UPDATE user_credentials SET role = $2 WHERE id = $1 RETURNING id `

	var updatedID string
	if err := s.db.QueryRowContext(ctx, query, userID, newRole).Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil
}

func (s *authStorage) Delete(ctx context.Context, userID string) error {

	query := `DELETE FROM user_credentials WHERE id=$1`

	if _, err := s.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	return nil
}

func (s *authStorage) ValidateLogin(ctx context.Context, username, password string) (bool, error) {

	var passwordHash string

	query := `SELECT password FROM user_credentials WHERE username=$1`
	err := s.db.QueryRowContext(ctx, query, username).Scan(&passwordHash)

	// invalid username
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	// invalid password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (s *authStorage) CheckUsernameExists(ctx context.Context, username string) (bool, error) {

	var exists bool

	query := `SELECT EXISTS (SELECT 1 FROM user_credentials WHERE username=$1);`
	if err := s.db.QueryRowContext(ctx, query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (s *authStorage) CheckExists(ctx context.Context, username, email, phoneNumber string) (string, error) {

	var existingField string

	row := s.db.QueryRowContext(ctx, `
		SELECT
			CASE
				WHEN username = $1 THEN 'username'
				WHEN email = $2 THEN 'email'
				WHEN phone_number = $3 THEN 'phone_number'
				ELSE NULL
			END AS existing_field
		FROM
			user_credentials
		WHERE
			username = $1 OR email = $2 OR phone_number = $3
		LIMIT 1;`, username, email, phoneNumber)

	if err := row.Scan(&existingField); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return existingField, nil
}

func hashPassword(password string) ([]byte, error) {
	hashedPass, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}
	return hashedPass, nil

}

package internal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	"golang.org/x/crypto/bcrypt"
)

type AuthStorage interface {

	// Users returns a list of user credentials.
	Users(ctx context.Context) ([]*dbUserCredentials, error)

	// User returns the user credentials.
	User(ctx context.Context, userID string) (*dbUserCredentials, error)

	// Create creates new user credential.
	Create(ctx context.Context, newUser *NewUserCredentials) (*dbUserCredentials, error)

	//Updates user role.
	UpdateRole(ctx context.Context, username string, newRole dbRoles) (*dbUserCredentials, error)

	// Deletes the user credential.
	Delete(ctx context.Context, userID string) error

	// Login validates username password and return user credentials
	Login(ctx context.Context, username, password string) (*dbUserCredentials, error)

	// Check user exists by username
	CheckUserExists(ctx context.Context, username string) (bool, error)
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

func (s *authStorage) Create(ctx context.Context, newUser *NewUserCredentials) (*dbUserCredentials, error) {

	hashedPass, err := hashPassword(newUser.Password)
	if err != nil {
		return nil, err
	}

	createTime := time.Now()

	res := s.db.QueryRowContext(ctx, `
		INSERT INTO user_credentials(
			username,
			email,
			password,
			role,
			phone_number,
			create_time
		)
		VALUES($1, $2, $3, $4,$5)
		RETURNING *;
	`,
		newUser.Username,
		newUser.Email,
		string(hashedPass),
		dbRoles(pb.Roles_USER),
		createTime,
	)

	var user dbUserCredentials
	err = res.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		nil,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
	)
	if err != nil {
		var pqError *pq.Error
		// 23505 = Unique constraint violation postgres
		if errors.As(err, &pqError) && pqError.Code == "23505" {
			return nil, errors.New("username or email duplicated")
		}
		return nil, err
	}

	return &user, nil
}

func (s *authStorage) UpdateRole(ctx context.Context, username string, newRole dbRoles) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		UPDATE 
			user_credentials
		SET 
			role = $1
		WHERE 
			username = $2
		RETURNING * ;
	`,
		newRole,
		username,
	)

	var user dbUserCredentials
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		nil,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (s *authStorage) Delete(ctx context.Context, userID string) error {

	_, err := s.db.ExecContext(ctx, `DELETE FROM user_credentials WHERE id=$1`, userID)
	if err != nil {
		return err
	}

	return nil
}
func (s *authStorage) Login(ctx context.Context, username, password string) (*dbUserCredentials, error) {

	var (
		user         dbUserCredentials
		passwordHash string
	)

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			id,
			username, 
			password,
			email,
			role
			phone_number,
			create_time
		FROM 
			user_credentials 
		WHERE 
			username=$1
	`,
		username,
	)
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&passwordHash,
		&user.Email,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
	)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, errors.New("username or password incorrect")
	}

	return &user, nil

}

func (s *authStorage) CheckUserExists(ctx context.Context, username string) (bool, error) {

	var user dbUserCredentials
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_credentials
			WHERE username=$1
		);
		`,
		username).Scan(&user)
	if err != nil {
		return false, err
	}

	return true, nil

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

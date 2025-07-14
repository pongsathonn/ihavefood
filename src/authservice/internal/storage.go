package internal

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *storage {
	return &storage{db: db}
}

func (s *storage) ListUsers(ctx context.Context) ([]*dbUserCredentials, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			user_id,
			username, 
			email,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials 
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
			&user.UpdateTime,
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

func (s *storage) GetUser(ctx context.Context, userID string) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			user_id,
			username, 
			email,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials
		WHERE
			user_id=$1
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
		&user.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (s *storage) GetUserByUsername(ctx context.Context, username string) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			user_id,
			username, 
			email,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials
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
		&user.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

// Create creates new user credential and return its ID.
func (s *storage) Create(ctx context.Context, newUser *NewUserCredentials) (string, error) {

	hashedPass, err := hashPassword(newUser.Password)
	if err != nil {
		return "", err
	}

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO credentials(
			username,
			email,
			password,
			role,
			phone_number,
			create_time
		)VALUES(
			$1,$2,$3,$4,$5,now()
		)RETURNING 
			user_id;
	`,
		newUser.Username,
		newUser.Email,
		string(hashedPass),
		newUser.Role,
		newUser.PhoneNumber,
	)

	var userID string
	if err := row.Scan(&userID); err != nil {
		return "", err
	}

	return userID, nil
}

// UpdateRole updates user role and returns ID.
func (s *storage) UpdateRole(ctx context.Context, userID string, newRole dbRoles) (string, error) {

	query := `UPDATE credentials SET role = $2 WHERE user_id = $1 RETURNING user_id `

	var updatedID string
	if err := s.db.QueryRowContext(ctx, query, userID, newRole).Scan(&updatedID); err != nil {
		return "", err
	}

	return updatedID, nil
}

// Delete deletes the user credential.
func (s *storage) Delete(ctx context.Context, userID string) error {

	query := `DELETE FROM credentials WHERE user_id=$1`

	if _, err := s.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	return nil
}

// ValidateLogin checks if the provided username and password are correct.
// It returns true if the credentials are valid, or false if they are not.
// Any non-nil error returned is related to the query execution not the
// validation process. Callers should check the error before interpreting
// the boolean result.
func (s *storage) ValidateLogin(ctx context.Context, username, password string) (bool, error) {

	var passwordHash string

	query := `SELECT password FROM credentials WHERE username=$1`
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

// CheckUsernameExists checks if the provided username already exists.
// It returns true if the username exists, false if it does not exist,
// and an error if any issues occur during the query process.
func (s *storage) CheckUsernameExists(ctx context.Context, username string) (bool, error) {

	query := `SELECT EXISTS (SELECT 1 FROM credentials WHERE username=$1);`

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
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

package internal

import (
	"context"
	"database/sql"
)

type storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *storage {
	return &storage{db: db}
}

func (s *storage) Begin() (*sql.Tx, error) {
	return s.db.Begin()
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
			&user.ID,
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
		&user.ID,
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

func (s *storage) GetUserByIdentifier(ctx context.Context, iden string) (*dbUserCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			user_id,
			username, 
			email,
			password,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials
		WHERE
			username=$1 OR
			email=$1 OR
			phone_number=$1 
	`,
		iden)

	var user dbUserCredentials
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPass,
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
func (s *storage) Create(ctx context.Context, newUser *dbNewUserCredentials) (*dbUserCredentials, error) {

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
		)RETURNING *;
	`,
		newUser.Username,
		newUser.Email,
		newUser.HashedPass,
		newUser.Role,
		newUser.PhoneNumber,
	)

	var user dbUserCredentials
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
		&user.UpdateTime,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

// Create creates new user credential and return its ID.
func (s *storage) CreateTx(ctx context.Context, tx *sql.Tx, newUser *dbNewUserCredentials) (*dbUserCredentials, error) {

	row := tx.QueryRowContext(ctx, `
		INSERT INTO credentials(
			username,
			email,
			password,
			role,
			phone_number,
			create_time
		)VALUES(
			$1,$2,$3,$4,$5,now()
		)RETURNING *;
	`,
		newUser.Username,
		newUser.Email,
		newUser.HashedPass,
		newUser.Role,
		newUser.PhoneNumber,
	)

	var user dbUserCredentials
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.PhoneNumber,
		&user.CreateTime,
		&user.UpdateTime,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete deletes the user credential.
func (s *storage) Delete(ctx context.Context, userID string) error {

	query := `DELETE FROM credentials WHERE user_id=$1`

	if _, err := s.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}

	return nil
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

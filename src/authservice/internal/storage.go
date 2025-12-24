package internal

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

func (s *storage) ListAuths(ctx context.Context) ([]*dbAuthCredentials, error) {

	rows, err := s.db.QueryContext(ctx, `
		SELECT 
			auth_id,
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

	var auths []*dbAuthCredentials
	for rows.Next() {
		var auth dbAuthCredentials
		err := rows.Scan(
			&auth.ID,
			&auth.Email,
			&auth.Role,
			&auth.PhoneNumber,
			&auth.CreateTime,
			&auth.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		auths = append(auths, &auth)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return auths, nil
}

func (s *storage) GetAuth(ctx context.Context, authID uuid.UUID) (*dbAuthCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			auth_id,
			email,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials
		WHERE
			auth_id=$1
	`,
		authID)

	var auth dbAuthCredentials
	err := row.Scan(
		&auth.ID,
		&auth.Email,
		&auth.Role,
		&auth.PhoneNumber,
		&auth.CreateTime,
		&auth.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	return &auth, nil

}

func (s *storage) GetAuthByIdentifier(ctx context.Context, iden string) (*dbAuthCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		SELECT 
			id,
			email,
			password,
			role,
			phone_number,
			create_time,
			update_time
		FROM 
			credentials
		WHERE
			email=$1 OR
			phone_number=$1 
	`,
		iden)

	var auth dbAuthCredentials
	err := row.Scan(
		&auth.ID,
		&auth.Email,
		&auth.HashedPass,
		&auth.Role,
		&auth.PhoneNumber,
		&auth.CreateTime,
		&auth.UpdateTime,
	)
	if err != nil {
		return nil, err
	}

	return &auth, nil

}

// Create creates new auth credential and return its ID.
func (s *storage) Create(ctx context.Context, newAuth *dbNewAuthCredentials) (*dbAuthCredentials, error) {

	row := s.db.QueryRowContext(ctx, `
		INSERT INTO credentials(
			email,
			password,
			role,
			phone_number
		)VALUES(
			$1,$2,$3,$4
		)RETURNING *;
	`,
		newAuth.Email,
		newAuth.HashedPass,
		newAuth.Role,
		newAuth.PhoneNumber,
	)

	var auth dbAuthCredentials
	if err := row.Scan(
		&auth.ID,
		&auth.Email,
		&auth.HashedPass,
		&auth.Role,
		&auth.PhoneNumber,
		&auth.CreateTime,
		&auth.UpdateTime,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, err
	}

	return &auth, nil
}

// Create creates new auth credential and return its ID.
func (s *storage) CreateTx(ctx context.Context, tx *sql.Tx, newAuth *dbNewAuthCredentials) (*dbAuthCredentials, error) {

	row := tx.QueryRowContext(ctx, `
		INSERT INTO credentials(
			email,
			password,
			role,
			phone_number,
			create_time
		)VALUES(
			$1,$2,$3,$4,now()
		)RETURNING *;
	`,
		newAuth.Email,
		newAuth.HashedPass,
		newAuth.Role,
		newAuth.PhoneNumber,
	)

	var auth dbAuthCredentials
	if err := row.Scan(
		&auth.ID,
		&auth.Email,
		&auth.HashedPass,
		&auth.Role,
		&auth.PhoneNumber,
		&auth.CreateTime,
		&auth.UpdateTime,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, err
	}

	return &auth, nil
}

// Delete deletes the auth credential.
func (s *storage) Delete(ctx context.Context, authID uuid.UUID) error {

	query := `DELETE FROM credentials WHERE auth_id=$1`

	if _, err := s.db.ExecContext(ctx, query, authID); err != nil {
		return err
	}

	return nil
}

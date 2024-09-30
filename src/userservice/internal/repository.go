package internal

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

type UserRepository interface {
	SaveUserProfile(ctx context.Context, username, phoneNumber string, address *address) (userId string, err error)
	UserProfiles(ctx context.Context) ([]*userProfile, error)
	UserProfile(ctx context.Context, userId string) (*userProfile, error)
	DeleteUserProfile(ctx context.Context, userId string) error
}

type userProfile struct {
	userId      string
	username    string
	phoneNumber string
	address     address
}

type address struct {
	addressName sql.NullString
	subDistrict sql.NullString
	district    sql.NullString
	province    sql.NullString
	postalCode  sql.NullString
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SaveUserProfile(ctx context.Context, username, phoneNumber string, address *address) (userId string, err error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}

	var addressID int
	addrRow := tx.QueryRowContext(ctx, `
		INSERT INTO address (
			name,
			sub_district,
			district,
			province,
			postal_code
		)
	    VALUES($1,$2,$3,$4,$5) 
		RETURNING id
	`,
		address.addressName,
		address.subDistrict,
		address.district,
		address.province,
		address.postalCode,
	)
	if err := addrRow.Scan(&addressID); err != nil {
		tx.Rollback()
		return "", err
	}

	var id int
	userRow := tx.QueryRowContext(ctx, `
		INSERT INTO user_profile(
			username,
			phone_number,
			address_id,
			created_at
		)
		VALUES($1,$2,$3,NOW())
		RETURNING id
	`,
		username,
		phoneNumber,
		addressID,
	)
	if err := userRow.Scan(&id); err != nil {
		tx.Rollback()
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	userID := strconv.Itoa(id)

	return userID, nil

}

func (r *userRepository) UserProfiles(ctx context.Context) ([]*userProfile, error) {

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			user_profile.id, 
			user_profile.username, 
			user_profile.phone_number,
			address.name, 
			address.sub_district, 
			address.district, 
			address.province, 
			address.postal_code
		FROM
			user_profile
		LEFT JOIN 
			address 
		ON 
			user_profile.address_id = address.id;
	`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	defer rows.Close()

	var users []*userProfile
	for rows.Next() {
		var user *userProfile
		err := rows.Scan(
			&user.userId,
			&user.username,
			&user.phoneNumber,
			&user.address.addressName,
			&user.address.subDistrict,
			&user.address.district,
			&user.address.province,
			&user.address.postalCode,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) UserProfile(ctx context.Context, userId string) (*userProfile, error) {

	row := r.db.QueryRowContext(ctx, `
		SELECT
			user_profile.id, 
			user_profile.username, 
			user_profile.phone_number,
			address.name, 
			address.sub_district, 
			address.district, 
			address.province, 
			address.postal_code
		FROM
			user_profile
		LEFT JOIN 
			address 
		ON 
			user_profile.address_id = address.id
		WHERE 
			user_profile.id = $1;
	`, userId)

	var user *userProfile
	err := row.Scan(
		&user.userId,
		&user.username,
		&user.phoneNumber,
		&user.address.addressName,
		&user.address.subDistrict,
		&user.address.district,
		&user.address.province,
		&user.address.postalCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) DeleteUserProfile(ctx context.Context, userId string) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	addressId, err := tx.ExecContext(ctx, `
		DELETE FROM 
			user_profile
		WHERE 
			id=$1
		RETURNNING 
			user_profile.address_id
	`, userId)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM address WHERE id=$1`, addressId); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

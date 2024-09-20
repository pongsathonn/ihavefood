package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository interface {
	SaveUserProfile(ctx context.Context, username, phoneNumber string, address *address) (string, error)
	UserProfiles(ctx context.Context) ([]*userProfile, error)
	UserProfile(ctx context.Context, username string) (*userProfile, error)
	DeleteUserProfile(ctx context.Context, username string) error
}

// TODO improve doc
// userProfile use for scan user from database
// it has sql.Null string for empty field in postgres
// proto.UserProfile does not have this
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

func (r *userRepository) SaveUserProfile(ctx context.Context, username, phoneNumber string, address *address) (string, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return "", status.Errorf(codes.Internal, "Failed to create user")
	}

	// Insert the address first
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
		log.Printf("Failed to insert address: %v", err)
		return "", status.Errorf(codes.Internal, "Failed to create user")
	}

	// Insert the user with the address ID
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
		log.Printf("Failed to insert user: %v", err)
		return "", status.Errorf(codes.Internal, "Failed to create user")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return "", status.Errorf(codes.Internal, "Failed to create user")
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
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.NotFound, "failed to retrive users from db")
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
			log.Printf("failed to scan: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to retrive users")
		}

		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) UserProfile(ctx context.Context, username string) (*userProfile, error) {

	if username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

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
			user_profile.username = $1;
	`, username)

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
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Printf("Scan failed :%v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive users from db")
	}
	return user, nil
}

func (r *userRepository) DeleteUserProfile(ctx context.Context, username string) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	addressId, err := tx.ExecContext(ctx, `
		DELETE FROM 
			user_profile
		WHERE 
			username=$1
		RETURNNING 
			user_profile.address_id
	`, username)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM address WHERE id=$1`, addressId)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

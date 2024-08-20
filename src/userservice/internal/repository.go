package internal

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

type UserRepository interface {
	SaveUserProfile(ctx context.Context, username, email, phoneNumber string, address *pb.Address) (string, error)
	UserProfiles(ctx context.Context) ([]*pb.UserProfile, error)
	GetUserProfileByUsername(ctx context.Context, username string) (*pb.UserProfile, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SaveUserProfile(ctx context.Context, username, email, phoneNumber string, address *pb.Address) (string, error) {

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
		address.AddressName,
		address.SubDistrict,
		address.District,
		address.Province,
		address.PostalCode,
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
			email,
			phone_number,
			address_id,
			created_at
		)
		VALUES($1,$2,$3,$4,NOW())
		RETURNING id
	`,
		username,
		email,
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

func (r *userRepository) UserProfiles(ctx context.Context) ([]*pb.UserProfile, error) {

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			user_profile.id, 
			user_profile.username, 
			user_profile.email, 
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

	var users []*pb.UserProfile
	for rows.Next() {
		var u userProfile
		err := rows.Scan(
			&u.UserId,
			&u.Username,
			&u.Email,
			&u.PhoneNumber,
			&u.AddressName,
			&u.SubDistrict,
			&u.District,
			&u.Province,
			&u.PostalCode,
		)
		if err != nil {
			log.Printf("failed to scan: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to retrive users")
		}

		user := &pb.UserProfile{
			UserId:      u.UserId,
			Username:    u.Username,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
			Address: &pb.Address{
				AddressName: u.AddressName.String,
				SubDistrict: u.SubDistrict.String,
				District:    u.District.String,
				Province:    u.Province.String,
				PostalCode:  u.PostalCode.String,
			},
		}

		users = append(users, user)
	}

	return users, nil

}

func (r *userRepository) GetUserProfileByUsername(ctx context.Context, username string) (*pb.UserProfile, error) {

	if username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	var user pb.UserProfile
	row := r.db.QueryRowContext(ctx, `
		SELECT
			user_profile.id, 
			user_profile.username, 
			user_profile.email, 
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
	if err := row.Scan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrive users from db")
	}

	return &user, nil
}

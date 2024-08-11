package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

type userProfile struct {
	UserId      string
	Username    string
	Email       string
	PhoneNumber string
	AddressName sql.NullString
	SubDistrict sql.NullString
	District    sql.NullString
	Province    sql.NullString
	PostalCode  sql.NullString
}

// userService handle user profiles
type userService struct {
	pb.UnimplementedUserServiceServer

	db *sql.DB
}

func NewUserService(db *sql.DB) *userService {
	return &userService{db: db}
}

func (x *userService) UpdateUser(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {

	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}

func (x *userService) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	if in.Username == "" || in.Email == "" || in.PhoneNumber == "" || in.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "username, email or phone number must be provided")
	}

	tx, err := x.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	// Insert the address first
	var addressID int
	addressQuery := `INSERT INTO address(name, sub_district, district, province, postal_code) 
	                 VALUES($1, $2, $3, $4, $5) RETURNING id`
	addrRow := tx.QueryRowContext(ctx, addressQuery,
		in.Address.AddressName,
		in.Address.SubDistrict,
		in.Address.District,
		in.Address.Province,
		in.Address.PostalCode,
	)
	if err := addrRow.Scan(&addressID); err != nil {
		tx.Rollback()
		log.Printf("Failed to insert address: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	// Insert the user with the address ID
	var id int
	userQuery := `INSERT INTO user_profile(username, email, phone_number, address_id, created_at) 
	              VALUES($1, $2, $3, $4, NOW()) RETURNING id`
	userRow := tx.QueryRowContext(ctx, userQuery,
		in.Username,
		in.Email,
		in.PhoneNumber,
		addressID,
	)
	if err := userRow.Scan(&id); err != nil {
		tx.Rollback()
		log.Printf("Failed to insert user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	userID := strconv.Itoa(id)

	return &pb.CreateUserResponse{UserId: userID}, nil
}

func (x *userService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {

	query :=
		`SELECT 
			user_profile.id, user_profile.username, user_profile.email, user_profile.phone_number,
			address.address_name, address.sub_district, address.district, address.province, address.postal_code
		FROM 
			user_profile
		LEFT JOIN address ON user_profile.address_id = address.id;`

	rows, err := x.db.QueryContext(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "")
	}
	defer rows.Close()

	var users []*pb.User

	for rows.Next() {
		var u userProfile
		if err := rows.Scan(&u); err != nil {
			log.Printf("Failed to scan: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to retrive users")
		}

		user := &pb.User{
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

	return &pb.ListUserResponse{Users: users}, nil
}

func (x *userService) GetUser(context.Context, *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

func (x *userService) DeleteUser(context.Context, *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

//TODO when user register , auth service or gateway need to send username and email via amqp to userService to save user

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

	db       *sql.DB
	rabbitmq RabbitmqClient
}

func NewUserService(db *sql.DB, rabbitmq RabbitmqClient) *userService {
	return &userService{db: db, rabbitmq: rabbitmq}
}

func (x *userService) UpdateUser(ctx context.Context, empty *pb.Empty) (*pb.Empty, error) {

	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}

func (x *userService) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	//TODO split this to another function
	// TODO create Datalayer to save this
	deliveries, err := x.rabbitmq.Subscribe("user_exchange", "user", "user.registered.event")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	var userx pb.CreateUserRequest
	for delivery := range deliveries {
		if err := json.Unmarshal(delivery.Body, &userx); err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

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

	//TODO validate input

	query :=
		`SELECT 
			user_profile.id, user_profile.username, user_profile.email, user_profile.phone_number,
			address.name, address.sub_district, address.district, address.province, address.postal_code
		FROM 
			user_profile
		LEFT JOIN address ON user_profile.address_id = address.id;`

	rows, err := x.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.NotFound, "failed to retrive users from db")
	}
	defer rows.Close()

	var users []*pb.User
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

func (x *userService) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	query :=
		`SELECT 
			user_profile.id, user_profile.username, user_profile.email, user_profile.phone_number,
			address.name, address.sub_district, address.district, address.province, address.postal_code
		FROM 
			user_profile
		LEFT JOIN address ON user_profile.address_id = address.id
		WHERE user_profile.username = $1;`

	var user pb.User
	if err := x.db.QueryRowContext(ctx, query, in.Username).Scan(&user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to retrive users from db")
	}

	return &pb.GetUserResponse{User: &user}, nil
}

func (x *userService) DeleteUser(context.Context, *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

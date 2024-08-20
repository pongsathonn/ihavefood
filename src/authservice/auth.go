package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

// auth implements the pb.AuthServiceServer interface.
type authService struct {
	pb.UnimplementedAuthServiceServer

	db         *sql.DB
	rabbitmq   RabbitmqClient
	userClient pb.UserServiceClient
}

// NewAuth creates a new instance of auth with the provided database connection.
func NewAuthService(db *sql.DB, rabbitmq RabbitmqClient, userClient pb.UserServiceClient) *authService {
	return &authService{
		db:         db,
		rabbitmq:   rabbitmq,
		userClient: userClient,
	}
}

// IsValidToken checks if the provided token is valid. It returns a response indicating validity and an error if any.
func (x *authService) IsValidToken(ctx context.Context, in *pb.IsValidTokenRequest) (*pb.IsValidTokenResponse, error) {

	if in.Token == "" {
		return nil, status.Errorf(codes.Unknown, "token must be provided")
	}

	if valid, err := validateToken(in.Token, signingKey); !valid {
		return nil, status.Errorf(codes.Unauthenticated, "token invalid: %v", err)
	}
	return &pb.IsValidTokenResponse{IsValid: true}, nil
}

// Register handles user registration. It creates a new user record with hashed password in the database.
// It returns an empty response on success or an error if the registration fails.
func (x *authService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username, email, or password must be provided")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "password hashing failed")
	}

	tx, err := x.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to begin transaction")
	}

	_, err = x.db.Exec(`
		INSERT INTO auth_table(
			username, 
			email, 
			password
		) 
		VALUES($1, $2, $3)
	`,
		in.Username,
		in.Email,
		string(hashedPass),
	)
	if err != nil {
		tx.Rollback()
		var pqError *pq.Error
		// 23505 = Unique constraint violation postgres
		if errors.As(err, &pqError) && pqError.Code == "23505" {
			return nil, status.Errorf(codes.AlreadyExists, "username or email duplicated ")
		}
		log.Printf("Failed to insert user into database: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert user into database")
	}

	// Calling UserService to update UserProfile
	req := &pb.CreateUserProfileRequest{
		Username:    in.Username,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Address:     in.Address,
	}
	_, err = x.userClient.CreateUserProfile(ctx, req)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to create user profile in UserService: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create user profile in UserService")
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to commit transaction")
	}

	return &pb.RegisterResponse{SuccessMessage: "registerd success"}, nil
}

// Login handles user login. It verifies the provided credentials, generates a JWT token on success, and returns it along with its expiration time.
// It returns an error if login fails or credentials are incorrect.
func (x *authService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user pb.UserCredentials
	row := x.db.QueryRowContext(ctx, `
		SELECT 
			username, 
			password 
		FROM 
			auth_table 
		WHERE 
			username=$1
	`,
		in.Username,
	)
	if err := row.Scan(&user.Username, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "scan failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "username or password incorrect")
	}

	token, exp, err := createNewToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate authentication token: %v", err)
	}

	// TODO: Publish event for user login

	return &pb.LoginResponse{AccessToken: token, AccessTokenExp: exp}, nil
}

// validateToken verifies the validity of a JWT token using the provided signing key.
// It returns true if the token is valid, false otherwise, along with any error encountered.
func validateToken(tokenString string, key []byte) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, fmt.Errorf("token invalid")
	}

	return true, nil
}

// createNewToken generates a new JWT token with a default expiration time of 5 minutes from the current time.
// It returns the signed token string, its expiration time in Unix format, and any error encountered.
func createNewToken() (string, int64, error) {

	// 1800 sec = 30 minutes
	addTimeSec := 1800
	unixNow := time.Now().Unix()
	expiration := unixNow + int64(addTimeSec)

	claims := &jwt.RegisteredClaims{
		Subject:   "authentication",
		Issuer:    "auth service",
		IssuedAt:  jwt.NewNumericDate(time.Unix(unixNow, 0)),
		ExpiresAt: jwt.NewNumericDate(time.Unix(expiration, 0)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %v", err)
	}

	return ss, expiration, nil
}

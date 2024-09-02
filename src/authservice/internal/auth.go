package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var signingKey []byte

type AuthClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	db         *sql.DB
	rabbitmq   RabbitMQ
	userClient pb.UserServiceClient
}

func NewAuthService(db *sql.DB, rabbitmq RabbitMQ, userClient pb.UserServiceClient) *AuthService {
	return &AuthService{
		db:         db,
		rabbitmq:   rabbitmq,
		userClient: userClient,
	}
}

func InitSigningKey() error {
	key := os.Getenv("JWT_SIGNING_KEY")
	if key == "" {
		return fmt.Errorf("JWT_SIGNING_KEY environment variable is empty")

	}
	signingKey = []byte(key)
	return nil
}

// initAdminUser create default user role (admin)
func InitAdminUser(db *sql.DB) error {

	admin := os.Getenv("INIT_ADMIN_USER")
	email := os.Getenv("INIT_ADMIN_EMAIL")
	password := os.Getenv("INIT_ADMIN_PASS")

	if admin == "" || email == "" || password == "" {
		return fmt.Errorf("required environment variables are not set")
	}

	hashedPass, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("password hashing failed")
	}

	_, err = db.Exec(`
		INSERT INTO user_credentials (
			username,
			email,
			password,
			role
		) 
		VALUES ($1,$2,$3,'ADMIN');
	`,
		admin,
		email,
		hashedPass,
	)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %v", err)
	}

	return nil
}

// Register handles user registration by creating a new user record with a hashed password in the database
// and calling the UserService to create a user profile. It uses a transaction to ensure that both the user
// creation and profile creation are successful before committing. If any error occurs, the transaction is rolled
// back to maintain data integrity. Returns a success response if all operations complete successfully, or an
// appropriate error if any operation fails.
func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username, email, or password must be provided")
	}

	hashedPass, err := hashPassword(in.Password)
	if err != nil {
		log.Println("Hasing failed: ", err)
		return nil, status.Errorf(codes.Internal, "password hashing failed")
	}

	tx, err := x.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to begin transaction")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_credentials(
			username,
			email,
			password,
			role
		)
		VALUES($1, $2, $3, $4)
	`,
		in.Username,
		in.Email,
		string(hashedPass),
		pb.Roles_USER.String(),
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

	// Calling UserService to create new UserProfile
	req := &pb.CreateUserProfileRequest{
		Username:    in.Username,
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

	return &pb.RegisterResponse{Success: true}, nil
}

// Login handles user login. It verifies the provided credentials, generates a JWT token on success, and returns it along with its expiration time.
// It returns an error if login fails or credentials are incorrect.
func (x *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user pb.UserCredentials
	row := x.db.QueryRowContext(ctx, `
		SELECT 
			username, 
			password,
			role
		FROM 
			user_credentials 
		WHERE 
			username=$1
	`,
		in.Username,
	)
	if err := row.Scan(&user.Username, &user.Password, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "scan failed: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "username or password incorrect")
	}

	token, exp, err := createNewToken(user.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate authentication token: %v", err)
	}

	// TODO: Publish event for user login

	return &pb.LoginResponse{AccessToken: token, AccessTokenExp: exp}, nil
}

// IsValidToken checks if the provided token is valid. It returns a response indicating validity and an error if any.
func (x *AuthService) IsValidToken(ctx context.Context, in *pb.IsValidTokenRequest) (*pb.IsValidTokenResponse, error) {

	if in.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token must be provided")
	}

	if valid, err := validateToken(in.Token); !valid {
		log.Println(err)
		return nil, status.Errorf(codes.Unauthenticated, "token invalid")
	}
	return &pb.IsValidTokenResponse{IsValid: true}, nil
}

func (x *AuthService) IsValidAdminToken(ctx context.Context, in *pb.IsValidAdminTokenRequest) (*pb.IsValidAdminTokenResponse, error) {

	if in.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token must be provided")
	}

	if valid, err := validateAdminToken(in.Token); !valid {
		return nil, status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
	}
	return &pb.IsValidAdminTokenResponse{IsValid: true}, nil

}

// FIXME might change this fn name from AssignRoles to UpdateUserRole
func (x *AuthService) AssignRolesToUsers(ctx context.Context, in *pb.AssignRolesToUsersRequest) (*pb.AssignRolesToUsersResponse, error) {

	if in.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username must be provided")
	}

	// check roles valid
	if _, ok := pb.Roles_value[in.Role.String()]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid roles")
	}

	_, err := x.db.ExecContext(ctx, `
		UPDATE user_credentials
		SET role = $1,
		WHERE username = $2
	`,
		in.Role.String(),
		in.Username,
	)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to update user into database")
	}

	// - update user role in database , mighe be create function update role

	return nil, status.Errorf(codes.Unimplemented, "method AssignRolesToUsers not implemented")
}

// validateToken verifies the validity of a JWT token using the provided signing key.
// It returns true if the token is valid, false otherwise, along with any error encountered.
func validateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, fmt.Errorf("token invalid")
	}

	return true, nil
}

func validateAdminToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, new(AuthClaims), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return false, err
	}

	if claims, _ := token.Claims.(*AuthClaims); claims.Role != pb.Roles_ADMIN.String() {
		return false, fmt.Errorf("Role invalid")
	}

	return true, nil
}

// createNewToken generates a new JWT token with a default expiration time of 5 minutes from the current time.
// It returns the signed token string, its expiration time in Unix format, and any error encountered.
//
// TODO modify createNewToken to handle both User and Admin
// by Fetch User role first and assign to claim
// modify doc to create new token based on role
func createNewToken(role pb.Roles) (string, int64, error) {

	// 1800 sec = 30 minutes
	addTimeSec := 1800
	unixNow := time.Now().Unix()
	expiration := unixNow + int64(addTimeSec)

	claims := &AuthClaims{
		Role: role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			Issuer:    "auth service",
			IssuedAt:  jwt.NewNumericDate(time.Unix(unixNow, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiration, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %v", err)
	}

	return ss, expiration, nil
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

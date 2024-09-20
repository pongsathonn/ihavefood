package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var signingKey []byte

var (
	errNoUsername         = status.Errorf(codes.InvalidArgument, "username must be provided")
	errNoPassword         = status.Errorf(codes.InvalidArgument, "password must be provided")
	errNoEmail            = status.Errorf(codes.InvalidArgument, "email must be provided")
	errNoUsernamePassword = status.Errorf(codes.InvalidArgument, "username or password must be provided")

	errUserIncorrect   = status.Errorf(codes.InvalidArgument, "username or password incorrect")
	errUserNotFound    = status.Errorf(codes.NotFound, "user not found")
	errPasswordHashing = status.Errorf(codes.Internal, "password hashing failed")

	errNoToken       = status.Errorf(codes.InvalidArgument, "token must be provided")
	errInvalidToken  = status.Errorf(codes.Unauthenticated, "invalid token")
	errGenerateToken = status.Errorf(codes.Internal, "failed to generate authentication token")
)

type AuthClaims struct {
	Role pb.Roles `json:"role"`
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
		return errors.New("JWT_SIGNING_KEY environment variable is empty")

	}
	signingKey = []byte(key)
	return nil
}

// InitAdminUser creates the default admin user if it doesn't already exist.
func InitAdminUser(db *sql.DB) error {

	admin := os.Getenv("INIT_ADMIN_USER")
	email := os.Getenv("INIT_ADMIN_EMAIL")
	password := os.Getenv("INIT_ADMIN_PASS")

	if admin == "" || email == "" || password == "" {
		return errors.New("required environment variables are not set")
	}

	// Check if the admin user already exists by username or email
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_credentials 
			WHERE username = $1 OR email = $2
		)
	`, admin, email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if admin user exists: %v", err)
	}

	if exists {
		return nil
	}

	// Hash the password
	hashedPass, err := hashPassword(password)
	if err != nil {
		log.Println(err)
		return errPasswordHashing
	}

	// Insert the admin user
	_, err = db.Exec(`
		INSERT INTO user_credentials (
			username,
			email,
			password,
			role
		) 
		VALUES ($1, $2, $3, 2);
	`,
		admin,
		email,
		hashedPass,
	)
	if err != nil {
		return fmt.Errorf("failed to insert admin user: %v", err)
	}

	log.Println("Admin user successfully initialized.")
	return nil
}

// TODO might not call userservice directly
//
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
		log.Printf("Hasing failed: %v", err)
		return nil, errPasswordHashing
	}

	tx, err := x.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to begin transaction")
	}
	defer tx.Rollback()

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
		pb.Roles_USER,
	)
	if err != nil {
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

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unknown, "missing metadata")
	}
	username, password, err := extractAuth(md["authorization"])
	if err != nil {
		log.Println("Invalid authorization: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization")
	}

	if username == "" || password == "" {
		return nil, errNoUsernamePassword
	}

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
		username,
	)
	if err := row.Scan(&user.Username, &user.Password, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errUserNotFound
		}
		log.Printf("Failed to scan: %v", err)
		return nil, status.Errorf(codes.Internal, "scan failed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errUserIncorrect
	}

	token, exp, err := createNewToken(user.Role)
	if err != nil {
		log.Println("Failed to generate token: %v", err)
		return nil, errGenerateToken
	}

	// TODO: Publish event for user login

	return &pb.LoginResponse{AccessToken: token, AccessTokenExp: exp}, nil
}

// IsValidToken checks if the provided token is valid. It returns a response indicating validity and an error if any.
func (x *AuthService) IsValidToken(ctx context.Context, in *pb.IsValidTokenRequest) (*pb.IsValidTokenResponse, error) {

	if in.Token == "" {
		return nil, errNoToken
	}

	if valid, err := validateToken(in.Token); !valid {
		log.Println(err)
		return nil, errInvalidToken
	}
	return &pb.IsValidTokenResponse{IsValid: true}, nil
}

func (x *AuthService) IsValidAdminToken(ctx context.Context, in *pb.IsValidAdminTokenRequest) (*pb.IsValidAdminTokenResponse, error) {

	if in.Token == "" {
		return nil, errNoToken
	}

	if valid, err := validateAdminToken(in.Token); !valid {
		log.Println("Token validation failed: %v", err)
		return nil, errInvalidToken
	}
	return &pb.IsValidAdminTokenResponse{IsValid: true}, nil
}

func (x *AuthService) IsUserExists(ctx context.Context, in *pb.IsUserExistsRequest) (*pb.IsUserExistsResponse, error) {
	if in.Username == "" {
		return nil, errNoUsername
	}

	var user *pb.UserCredentials
	err := x.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM user_credentials
			WHERE username=$1
		);
		`,
		in.Username).Scan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errUserNotFound
		}
	}

	// TODO response
	return nil, status.Errorf(codes.Unimplemented, "method IsUserExists not implemented")
}

func (x *AuthService) UpdateUserRole(ctx context.Context, in *pb.UpdateUserRoleRequest) (*pb.UpdateUserRoleResponse, error) {

	if in.Username == "" {
		return nil, errNoUsername
	}

	// check roles valid with comma ok
	if _, ok := pb.Roles_value[in.Role.String()]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "role %s invalid", in.Role.String())
	}

	_, err := x.db.ExecContext(ctx, `
		UPDATE user_credentials
		SET role = $1
		WHERE username = $2
	`,
		in.Role,
		in.Username,
	)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to update role")
	}

	// - update user role in database , mighe be create function update role

	return &pb.UpdateUserRoleResponse{Success: true}, nil
}

func extractAuth(authorization []string) (username, password string, err error) {

	if len(authorization) < 1 {
		return "", "", errors.New("missing authorization in metadata")
	}

	encoded := strings.TrimPrefix(authorization[0], "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}

	cred := strings.Split(string(decoded), ":")
	return cred[0], cred[1], nil
}

// createNewToken generates a new JWT token with a expiration time from the current time.
// It returns the signed token string, its expiration time in Unix format, and any error encountered.
//
// TODO modify createNewToken to handle both User and Admin
// by Fetch User role first and assign to claim
// modify doc to create new token based on role
func createNewToken(role pb.Roles) (string, int64, error) {

	day := 24 * time.Hour
	expiration := time.Now().Add(7 * day).Unix()

	claims := &AuthClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			Issuer:    "auth service",
			IssuedAt:  jwt.NewNumericDate(time.Unix(time.Now().Unix(), 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiration, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, err
	}

	return ss, expiration, nil
}

// TODO might be check "USER" role
//
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
		return false, errors.New("invalid token")
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

	if claims, _ := token.Claims.(*AuthClaims); claims.Role != pb.Roles_ADMIN {
		return false, errors.New("invalid role")
	}

	return true, nil
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

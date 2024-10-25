package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

var signingKey []byte

// TODO might move to error.go file
var (
	errNoUsername         = status.Error(codes.InvalidArgument, "username must be provided")
	errNoPassword         = status.Error(codes.InvalidArgument, "password must be provided")
	errNoEmail            = status.Error(codes.InvalidArgument, "email must be provided")
	errNoUsernamePassword = status.Error(codes.InvalidArgument, "username or password must be provided")

	errUserIncorrect   = status.Error(codes.InvalidArgument, "username or password incorrect")
	errUserNotFound    = status.Error(codes.NotFound, "user not found")
	errPasswordHashing = status.Error(codes.Internal, "password hashing failed")

	errNoToken       = status.Error(codes.InvalidArgument, "token must be provided")
	errInvalidToken  = status.Error(codes.Unauthenticated, "invalid token")
	errGenerateToken = status.Error(codes.Internal, "failed to generate authentication token")
)

type AuthClaims struct {
	ID   string
	Role pb.Roles `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	store      AuthStorage
	rabbitmq   RabbitMQ
	userClient pb.UserServiceClient
}

func NewAuthService(store AuthStorage, rabbitmq RabbitMQ, userClient pb.UserServiceClient) *AuthService {
	return &AuthService{
		store:      store,
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

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_credentials 
			WHERE username = $1 OR email = $2
		)
	`, admin, email).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	hashedPass, err := hashPassword(password)
	if err != nil {
		return err
	}

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
		return err
	}

	slog.Info("admin successfully innitialized")

	return nil
}

// Register handles user registration by creating a new user record with a hashed password
// in the database and calling the UserService to create a user profile. It uses a transaction
// to ensure that both the user creation and profile creation are successful before committing.
// If any error occurs, the transaction is rolled back to maintain data integrity. Returns
// a success response if all operations complete successfully, or an appropriate error
// if any operation fails.

func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.UserCredentials, error) {
	if err := validateUsernameAndPassword(in.Username, in.Password); err != nil {
		slog.Error("validate user", "err", err)
		return nil, status.Error(codes.InvalidArgument, "username or password invalid")
	}

	if err := validatePhoneNumber(in.PhoneNumber); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "phone number invalid: %v", err)
	}

	if err := validateEmail(in.Email); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email invalid: %v", err)
	}

	user, err := x.store.Create(ctx, &NewUserCredentials{
		Username:    in.Username,
		Email:       in.Email,
		Password:    in.Password,
		PhoneNumber: in.PhoneNumber,
	})

	// TODO remove UserService call. publish user credentials with event instead

	// Calling UserService to create new UserProfile
	_, err = x.userClient.CreateProfile(ctx,
		&pb.CreateProfileRequest{
			Username: user.Username,
		})
	if err != nil {
		slog.Error("failed to create user profile",
			"err", fmt.Errorf("error user service: %v", err),
		)

		// to ensure that AuthService and UserService syn
		if err := x.store.Delete(ctx, user.UserID); err != nil {
			slog.Error("failed to delete user credential", "err", err)
		}

		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.UserCredentials{
		UserId:      user.UserID,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        pb.Roles(user.Role),
	}, nil
}

// Login handles user login. It verifies the provided credentials, generates a JWT token on success,
// and returns it along with its expiration time. It returns an error if login fails or credentials are incorrect.
func (x *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "missing metadata")
	}
	username, password, err := extractBasicAuth(md["authorization"])
	if err != nil {
		slog.Error("authorization", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid authorization")
	}

	if username == "" || password == "" {
		return nil, errNoUsernamePassword
	}

	user, err := x.store.ValidateLogin(ctx, in.Username, in.Password)
	if err != nil {
		slog.Error("failed to validate user login", "err", err)
		return nil, errUserIncorrect
	}

	token, exp, err := createNewToken(user.UserID, pb.Roles(user.Role))
	if err != nil {
		slog.Error("generate new token", "err", err)
		return nil, errGenerateToken
	}

	return &pb.LoginResponse{AccessToken: token, AccessTokenExp: exp}, nil
}

// TODO might user "Validate" instead "Check"
func (x *AuthService) CheckUserToken(ctx context.Context, in *pb.CheckUserTokenRequest) (*pb.CheckUserTokenResponse, error) {

	if in.Token == "" {
		return nil, errNoToken
	}

	if valid, err := validateUserToken(in.Token); !valid {
		slog.Error("validate token", "err", err)
		return nil, errInvalidToken
	}
	return &pb.CheckUserTokenResponse{IsValid: true}, nil
}

func (x *AuthService) CheckAdminToken(ctx context.Context, in *pb.CheckAdminTokenRequest) (*pb.CheckAdminTokenResponse, error) {

	if in.Token == "" {
		return nil, errNoToken
	}

	if valid, err := validateAdminToken(in.Token); !valid {
		slog.Error("validate admin token", "err", err)
		return nil, errInvalidToken
	}
	return &pb.CheckAdminTokenResponse{IsValid: true}, nil
}

func (x *AuthService) CheckUserExists(ctx context.Context, in *pb.CheckUserExistsRequest) (*pb.CheckUserExistsResponse, error) {

	if in.Username == "" {
		return nil, errNoUsername
	}

	exists, err := x.store.CheckUserExists(ctx, in.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errUserNotFound
		}
		slog.Error("failed to check existence user", "err", err)
		return nil, status.Error(codes.Internal, "failed to check existence user")
	}

	if !exists {
		return nil, status.Error(codes.NotFound, "user not exists")
	}

	return &pb.CheckUserExistsResponse{IsExists: true}, nil
}

// UpdateUserRole updates an existing user's role to specific roles.
//
// NOTE: Calling this function should be preceded by middleware first to
// prevent lower roles updating highter roles.For exmaple "USER" try to
// update "ADMIN", which is invalid.
func (x *AuthService) UpdateUserRole(ctx context.Context, in *pb.UpdateUserRoleRequest) (*pb.UpdateUserRoleResponse, error) {

	if in.Username == "" {
		return nil, errNoUsername
	}

	if _, ok := pb.Roles_value[in.Role.String()]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "role %s invalid", in.Role.String())
	}

	if err := x.store.UpdateRole(ctx, in.Username, dbRoles(in.Role)); err != nil {
		slog.Error("failed to update role", "err", err)
		return nil, status.Error(codes.Internal, "failed to update role")
	}

	return &pb.UpdateUserRoleResponse{Success: true}, nil
}

func extractBasicAuth(authorization []string) (username, password string, err error) {

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

// createNewToken generates a new JWT token specific roles with an expiration time from the current time.
// It returns the signed token string, its expiration time in Unix format, and any error encountered.
func createNewToken(id string, role pb.Roles) (signedToken string, expiration int64, err error) {

	day := 24 * time.Hour
	exp := time.Now().Add(7 * day).Unix()

	claims := &AuthClaims{
		ID:   id,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			Issuer:    "auth service",
			IssuedAt:  jwt.NewNumericDate(time.Unix(time.Now().Unix(), 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(exp, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, err
	}

	return ss, exp, nil
}

// validateUserToken verifies the validity of a JWT token using the provided signing key.
// It returns true if the token is valid, false otherwise, along with any error encountered.
func validateUserToken(tokenString string) (bool, error) {
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
		return false, errors.New("invalid user token")
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
		return false, errors.New("token claims do not have admin role")
	}

	return true, nil
}

func validateUsernameAndPassword(username, password string) error {
	if username == "" || password == "" {
		return errors.New("username or password must be provided")
	}

	if len(username) < 6 || len(username) > 16 {
		return errors.New("username must be between 6 and 16 characters")
	}

	if len(password) < 8 || len(password) > 16 {
		return errors.New("password must be between 8 and 16 characters")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(username) {
		return errors.New("username can only contain letters and numbers")
	}

	if strings.TrimSpace(username) != username {
		return errors.New("username cannot start or end with spaces")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !regexp.MustCompile(`[!.@#$%^&*()_\-+=<>?]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// validateEmail validates the user's email address to ensure it follows
// the standard email format. It uses mail.ParseAddress to parse the email.
// If the email is invalid, it returns an error.
//
// NOTE email such as "test@gmailcom" (without dot) not counted as error
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

// validatePhoneNumber validates a user's phone number according to the Thailand
// phone number format (e.g., 06XXXXXXXX, 08XXXXXXXX, 09XXXXXXXX).
// Any format outside of this is considered invalid, and the function returns an error.
func validatePhoneNumber(phoneNumber string) error {
	if !regexp.MustCompile(`^(06|08|09)\d{8}$`).MatchString(phoneNumber) {
		return errors.New("invalid phone number format")
	}
	return nil
}

package internal

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	store         AuthStorage
	rabbitmq      RabbitMQ
	profileClient pb.ProfileServiceClient
}

func NewAuthService(store AuthStorage, rabbitmq RabbitMQ, profileClient pb.ProfileServiceClient) *AuthService {
	return &AuthService{
		store:         store,
		rabbitmq:      rabbitmq,
		profileClient: profileClient,
	}
}

// Register handles user registration by creating a new user credentials
// and calling the UserService to create a user profile.
func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.UserCredentials, error) {

	if err := validateUser(in); err != nil {
		slog.Error("failed to validate user", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "validation failed : %v", err)
	}

	userID, err := x.store.Create(ctx, &NewUserCredentials{
		Username:    in.Username,
		Email:       in.Email,
		Password:    in.Password,
		PhoneNumber: in.PhoneNumber,
		Role:        Roles_USER,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user %v", err)
	}

	user, err := x.store.User(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrive user %v", err)
	}

	// Creates new user profile in UserService
	if err := x.createUserProfile(user.UserID, user.Username); err != nil {

		slog.Error("user profile creation failed: %v", err)

		if err := x.store.Delete(context.TODO(), user.UserID); err != nil {
			slog.Error("failed to delete user: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.UserCredentials{
		UserId:      user.UserID,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        pb.Roles(user.Role),
		CreateTime:  timestamppb.New(user.CreateTime),
	}, nil
}

// Login handles user login. It verifies the provided credentials, generates a JWT token on success,
// and returns it along with its expiration time. It returns an error if login fails or credentials are incorrect.
func (x *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {

	// For Login with http "GET" method
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, status.Error(codes.Unknown, "missing metadata")
	// }

	// username, password, err := extractBasicAuth(md["authorization"])
	// if err != nil {
	// 	slog.Error("authorization", "err", err)
	// 	return nil, status.Error(codes.Unauthenticated, "invalid authorization")
	// }

	// TODO validate login request body

	valid, err := x.store.ValidateLogin(ctx, in.Username, in.Password)
	if err != nil {
		slog.Error("failed to validate user login", "err", err)
		return nil, status.Error(codes.Internal, "failed to validate user login")
	}

	if !valid {
		return nil, errUserIncorrect
	}

	user, err := x.store.UserByUsername(ctx, in.Username)
	if err != nil {
		slog.Error("failed to find user credentials", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to find user credentials %v", err)
	}

	token, exp, err := createNewToken(user.UserID, pb.Roles(user.Role))
	if err != nil {
		slog.Error("generate new token", "err", err)
		return nil, errGenerateToken
	}

	return &pb.LoginResponse{
		AccessToken: token,
		ExpiresIn:   exp,
	}, nil
}

func (x *AuthService) ValidateUserToken(ctx context.Context, in *pb.ValidateUserTokenRequest) (*pb.ValidateUserTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, errNoToken
	}

	if valid, err := validateUserToken(in.AccessToken); !valid {
		slog.Error("validate token", "err", err)
		return nil, errInvalidToken
	}
	return &pb.ValidateUserTokenResponse{Valid: true}, nil
}

func (x *AuthService) ValidateAdminToken(ctx context.Context, in *pb.ValidateAdminTokenRequest) (*pb.ValidateAdminTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, errNoToken
	}

	if valid, err := validateAdminToken(in.AccessToken); !valid {
		slog.Error("validate admin token", "err", err)
		return nil, errInvalidToken
	}
	return &pb.ValidateAdminTokenResponse{Valid: true}, nil
}

func (x *AuthService) CheckUsernameExists(ctx context.Context, in *pb.CheckUsernameExistsRequest) (*pb.CheckUsernameExistsResponse, error) {

	// TODO validate request

	exists, err := x.store.CheckUsernameExists(ctx, in.Username)
	if err != nil {
		slog.Error("failed to check existence user", "err", err)
		return nil, status.Error(codes.Internal, "failed to check existence user")
	}

	if !exists {
		return nil, errUserNotFound
	}

	return &pb.CheckUsernameExistsResponse{Exists: true}, nil
}

// UpdateUserRole updates an existing user's role to specific roles.
//
// NOTE: Calling this function should be preceded by middleware first to
// prevent lower roles updating highter roles.
func (x *AuthService) UpdateUserRole(ctx context.Context, in *pb.UpdateUserRoleRequest) (*pb.UserCredentials, error) {

	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "userID must be provided")
	}

	if _, ok := pb.Roles_value[in.NewRole.String()]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "role %s invalid", in.NewRole.String())
	}

	updatedID, err := x.store.UpdateRole(ctx, in.UserId, dbRoles(in.NewRole))
	if err != nil {
		slog.Error("failed to update user role", "err", err)
		return nil, status.Error(codes.Internal, "failed to update role")
	}

	user, err := x.store.User(ctx, updatedID)
	if err != nil {
		slog.Error("failed to find user credentials", "err", err)
		return nil, status.Error(codes.Internal, "failed to find user credentials")
	}

	//TODO add update_time

	return &pb.UserCredentials{
		UserId:      user.UserID,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        pb.Roles(user.Role),
		CreateTime:  timestamppb.New(user.CreateTime),
	}, nil
}

func (x *AuthService) createUserProfile(userID, username string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profile, err := x.profileClient.CreateProfile(ctx, &pb.CreateProfileRequest{
		UserId:   userID,
		Username: username,
	})
	if err != nil {
		return err
	}

	if profile.UserId != userID && profile.Username != username {
		return errors.New("userID or username mismatch")
	}

	return nil

}

func InitSigningKey() error {
	if key := os.Getenv("JWT_SIGNING_KEY"); key != "" {
		signingKey = []byte(key)
		return nil
	}
	return errors.New("JWT_SIGNING_KEY environment variable is empty")
}

// InitAdminUser creates the default admin user. The reason the default admin is created
// in Go is to ensure that the password is hashed using the same hashing function.
func InitAdminUser(storage AuthStorage) error {

	admin := os.Getenv("INIT_ADMIN_USER")
	email := os.Getenv("INIT_ADMIN_EMAIL")
	password := os.Getenv("INIT_ADMIN_PASS")

	if admin == "" || email == "" || password == "" {
		return errors.New("some of admin environment variables are not set")
	}

	if _, err := storage.Create(context.TODO(), &NewUserCredentials{
		Username: admin,
		Email:    email,
		Password: password,
		Role:     dbRoles(Roles_ADMIN),
	}); err != nil {
		return err
	}

	log.Println("admin successfully innitialized")

	return nil
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
	now := time.Now()
	exp := now.Add(30 * day)

	claims := &AuthClaims{
		ID:   id,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "authentication",
			Issuer:    "auth service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", 0, err
	}

	return ss, exp.Unix(), nil
}

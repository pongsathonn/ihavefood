package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

type AuthStorer interface {
	Begin() (*sql.Tx, error)
	ListUsers(context.Context) ([]*dbUserCredentials, error)
	GetUser(ctx context.Context, userID string) (*dbUserCredentials, error)
	GetUserByIdentifier(ctx context.Context, iden string) (*dbUserCredentials, error)
	Create(ctx context.Context, newUser *dbNewUserCredentials) (*dbUserCredentials, error)
	CreateTx(ctx context.Context, tx *sql.Tx, newUser *dbNewUserCredentials) (*dbUserCredentials, error)
	Delete(ctx context.Context, userID string) error
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
}

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	store          AuthStorer
	customerClient pb.CustomerServiceClient
	deliveryClient pb.DeliveryServiceClient
	merchantClient pb.MerchantServiceClient
}

type AuthCfg struct {
	Store          AuthStorer
	CustomerClient pb.CustomerServiceClient
	DeliveryClient pb.DeliveryServiceClient
	MerchantClient pb.MerchantServiceClient
}

func NewAuthService(cfg *AuthCfg) *AuthService {
	return &AuthService{
		store:          cfg.Store,
		customerClient: cfg.CustomerClient,
		deliveryClient: cfg.DeliveryClient,
		merchantClient: cfg.MerchantClient,
	}
}

// Register handles user registration by creating a new user credentials
// and calling the UserService to create a user customer.
func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.UserCredentials, error) {

	if err := validateUser(in); err != nil {
		slog.Error("failed to validate user", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "validation failed : %v", err)
	}

	switch in.Role {
	case pb.Roles_MERCHANT:
		if in.Username != "" {
			return nil, status.Errorf(codes.InvalidArgument, "merchant do not use usernames")
		}
	case pb.Roles_UNKNOWN, pb.Roles_ADMIN, pb.Roles_SUPER_ADMIN:
		return nil, status.Error(codes.InvalidArgument, "invalid role for registration")
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported role provided: %s", in.Role.String())
	}

	hashPass, err := hashPassword(in.Password)
	if err != nil {
		slog.Error("failed to hash password", "err", err)
		return nil, status.Error(codes.Internal, "hashing password failed")
	}

	tx, err := x.store.Begin()
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	defer tx.Rollback()

	user, err := x.store.CreateTx(ctx, tx, &dbNewUserCredentials{
		Username:    in.Username,
		Email:       in.Email,
		HashedPass:  string(hashPass),
		PhoneNumber: in.PhoneNumber,
		Role:        dbRoles(in.Role),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user %v", err)
	}

	switch in.Role {
	case pb.Roles_CUSTOMER:
		if err := x.createCustomer(ctx, user); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create user %v", err)
		}
	case pb.Roles_RIDER:
		// TODO: impl create rider
	case pb.Roles_MERCHANT:
		// TODO: impl create rider
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "unable to commit register transaction")
	}

	return &pb.UserCredentials{
		Id:          user.ID,
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

	// For login with HTTP GET
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, status.Error(codes.Unknown, "missing metadata")
	// }
	// username, password, err := extractBasicAuth(md["authorization"])
	// if err != nil {
	// 	slog.Error("authorization", "err", err)
	// 	return nil, status.Error(codes.Unauthenticated, "invalid authorization")
	// }

	user, err := x.store.GetUserByIdentifier(ctx, in.Identifier)
	if err != nil {
		slog.Error("failed to find user credentials", "err", err)
		return nil, status.Error(codes.Internal, "failed to find user credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPass), []byte(in.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errUserIncorrect
		}
		slog.Error("bcrypt verification failed unexpectedly", "err", err)
		return nil, status.Error(codes.Internal, "authentication failed due to server error")
	}

	token, exp, err := createNewToken(user.ID, pb.Roles(user.Role))
	if err != nil {
		slog.Error("generate new token", "err", err)
		return nil, errGenerateToken
	}

	return &pb.LoginResponse{
		AccessToken: token,
		ExpiresIn:   exp,
	}, nil
}

func (x *AuthService) VerifyUserToken(ctx context.Context, in *pb.VerifyUserTokenRequest) (*pb.VerifyUserTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, errNoToken
	}

	if valid, err := verifyUserToken(in.AccessToken); !valid {
		slog.Error("verify token", "err", err)
		return nil, errInvalidToken
	}

	return &pb.VerifyUserTokenResponse{Valid: true}, nil
}

func (x *AuthService) VerifyAdminToken(ctx context.Context, in *pb.VerifyAdminTokenRequest) (*pb.VerifyAdminTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, errNoToken
	}

	if valid, err := verifyAdminToken(in.AccessToken); !valid {
		slog.Error("verify admin token", "err", err)
		return nil, errInvalidToken
	}
	return &pb.VerifyAdminTokenResponse{Valid: true}, nil
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

func (x *AuthService) createCustomer(ctx context.Context, user *dbUserCredentials) error {
	customer, err := x.customerClient.CreateCustomer(ctx, &pb.CreateCustomerRequest{
		CustomerId: user.ID,
		Username:   user.Username,
	})
	if err != nil {
		return err
	}
	if user.ID != customer.CustomerId || user.Username != customer.Username {
		return errors.New("ID or Username in AuthService and CustomerService are inconsistent")
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

func CreateSuperAdmin(store AuthStorer) error {

	admin := os.Getenv("SUPER_ADMIN_USER")
	email := os.Getenv("SUPER_ADMIN_EMAIL")
	password := os.Getenv("SUPER_ADMIN_PASS")

	if admin == "" || email == "" || password == "" {
		return errors.New("some of super admin environment variables are not set")
	}

	hashPass, err := hashPassword(password)
	if err != nil {
		return errors.New("hashing password failed")
	}

	if _, err := store.Create(context.TODO(), &dbNewUserCredentials{
		Username:   admin,
		Email:      email,
		HashedPass: string(hashPass),
		Role:       dbRoles(Roles_SUPER_ADMIN),
	}); err != nil {
		return err
	}

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

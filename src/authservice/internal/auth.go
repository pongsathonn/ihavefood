package internal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

type AuthStorer interface {
	Begin() (*sql.Tx, error)
	ListUsers(context.Context) ([]*dbUserCredentials, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*dbUserCredentials, error)
	GetUserByIdentifier(ctx context.Context, iden string) (*dbUserCredentials, error)
	Create(ctx context.Context, newUser *dbNewUserCredentials) (*dbUserCredentials, error)
	CreateTx(ctx context.Context, tx *sql.Tx, newUser *dbNewUserCredentials) (*dbUserCredentials, error)
	Delete(ctx context.Context, userID uuid.UUID) error
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
func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.UserCredentials, error) {

	if err := ValidateStruct(in); err != nil {
		var ve myValidatorErrs
		if errors.As(err, &ve) {
			return nil, status.Errorf(codes.InvalidArgument, "failed to register: %s", ve.Error())
		}
		slog.Error("validate struct", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	hashPass, err := hashPassword(in.Password)
	if err != nil {
		slog.Error("hashing password", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	tx, err := x.store.Begin()
	if err != nil {
		slog.Error("begin transaction", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
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
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, errors.New("user already exists")
		}

		slog.Error("storage create new user", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := x.dispatchCreation(ctx, in.Role, user); err != nil {
		slog.Error("dispatch creation", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := tx.Commit(); err != nil {
		slog.Error("commit transaction", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.UserCredentials{
		Id:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        pb.Roles(user.Role),
		CreateTime:  timestamppb.New(user.CreateTime),
		UpdateTime:  timestamppb.New(user.UpdateTime),
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

	if err := ValidateStruct(in); err != nil {
		var ve myValidatorErrs
		if errors.As(err, &ve) {
			return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to login: %s", ve.Error()))
		}
		slog.Error("validate struct", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	user, err := x.store.GetUserByIdentifier(ctx, in.Identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("storage get user by identifier not found")
			return nil, status.Error(codes.Unauthenticated, "username or password incorrect")
		}
		slog.Error("storage get user by identifier", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPass), []byte(in.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			slog.Error("bcrypt error mismatch")
			return nil, status.Error(codes.Unauthenticated, "username or password incorrect")
		}
		slog.Error("bcrypt verification failed unexpectedly", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	token, exp, err := createNewToken(user.ID, pb.Roles(user.Role))
	if err != nil {
		slog.Error("create new token", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.LoginResponse{
		AccessToken: token,
		ExpiresIn:   exp,
	}, nil
}

func (x *AuthService) VerifyUserToken(ctx context.Context, in *pb.VerifyUserTokenRequest) (*pb.VerifyUserTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token must be provided")
	}

	if valid, err := verifyUserToken(in.AccessToken); !valid {
		slog.Error("verify user token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid user token")
	}

	return &pb.VerifyUserTokenResponse{Valid: true}, nil
}

func (x *AuthService) VerifyAdminToken(ctx context.Context, in *pb.VerifyAdminTokenRequest) (*pb.VerifyAdminTokenResponse, error) {

	if in.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token must be provided")
	}

	if valid, err := verifyAdminToken(in.AccessToken); !valid {
		slog.Error("verify admin token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid admin token")
	}
	return &pb.VerifyAdminTokenResponse{Valid: true}, nil
}

func (x *AuthService) dispatchCreation(ctx context.Context, role pb.Roles, user *dbUserCredentials) error {
	switch role {
	case pb.Roles_CUSTOMER:
		if err := x.createCustomer(ctx, user); err != nil {
			return fmt.Errorf("failed to create customer: %v", err)
		}
	case pb.Roles_RIDER:
		if err := x.createRider(ctx, user); err != nil {
			return fmt.Errorf("failed to create rider: %v", err)
		}
	case pb.Roles_MERCHANT:
		if err := x.createMerchant(ctx, user); err != nil {
			return fmt.Errorf("failed to create merchant: %v", err)
		}
	default:
		return errors.New("invalid role")
	}

	return nil
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

func (x *AuthService) createMerchant(ctx context.Context, user *dbUserCredentials) error {

	defaultMerchantName, _, found := strings.Cut(user.Email, "@")
	if !found {
		return errors.New("invalid email format: missing '@'")
	}

	merchant, err := x.merchantClient.CreateMerchant(ctx, &pb.CreateMerchantRequest{
		MerchantId:   user.ID,
		MerchantName: defaultMerchantName,
	})
	if err != nil {
		return err
	}
	if merchant.MerchantId != user.ID {
		return errors.New("ID in AuthService and MerchantService are inconsistent")
	}
	return nil
}

func (x *AuthService) createRider(ctx context.Context, user *dbUserCredentials) error {
	rider, err := x.deliveryClient.CreateRider(ctx, &pb.CreateRiderRequest{
		RiderId:     user.ID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
	})
	if err != nil {
		return err
	}
	if rider.RiderId != user.ID || rider.Username != user.Username {
		return errors.New("ID or Username in AuthService and DeliveryService are inconsistent")
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

// func statusErrInfo(c codes.Code, msg string, reason pb.Reason, meta map[string]string) error {
// 	st := status.New(c, msg)
// 	wd, err := st.WithDetails(&epb.ErrorInfo{
// 		Reason:   reason.String(),
// 		Domain:   pb.AuthService_ServiceDesc.ServiceName,
// 		Metadata: meta,
// 	})
// 	if err != nil {
// 		return st.Err()
// 	}
//
// 	return wd.Err()
// }

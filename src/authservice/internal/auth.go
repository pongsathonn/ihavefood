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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AuthStorer interface {
	Begin() (*sql.Tx, error)
	ListAuths(context.Context) ([]*dbAuthCredentials, error)
	GetAuth(ctx context.Context, authID uuid.UUID) (*dbAuthCredentials, error)
	GetAuthByIdentifier(ctx context.Context, iden string) (*dbAuthCredentials, error)
	Create(ctx context.Context, newAuth *dbNewAuthCredentials) (*dbAuthCredentials, error)
	CreateTx(ctx context.Context, tx *sql.Tx, newAuth *dbNewAuthCredentials) (*dbAuthCredentials, error)
	Delete(ctx context.Context, authID uuid.UUID) error
}

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	store    AuthStorer
	rabbitmq *rabbitMQ
}

func NewAuthService(store AuthStorer, rabbitmq *rabbitMQ) *AuthService {
	return &AuthService{store: store, rabbitmq: rabbitmq}
}

// Register handles auth registration by creating a new auth credentials
func (x *AuthService) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.AuthCredentials, error) {

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

	auth, err := x.store.CreateTx(ctx, tx, &dbNewAuthCredentials{
		Email:       in.Email,
		HashedPass:  string(hashPass),
		Role:        dbRoles(in.Role),
		PhoneNumber: nil,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			slog.Error("database unique_violation:", "err", err)
			return nil, status.Error(codes.AlreadyExists, "email or phone number already exists")
		}

		slog.Error("storage create new auth", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := x.dispatchCreation(ctx, in.Role, auth); err != nil {
		slog.Error("dispatch creation", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := tx.Commit(); err != nil {
		slog.Error("commit transaction", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")

	}

	phone := ""
	if auth.PhoneNumber != nil {
		phone = *auth.PhoneNumber
	}

	return &pb.AuthCredentials{
		Id:         auth.ID,
		Email:      auth.Email,
		Phone:      phone,
		Role:       pb.Roles(auth.Role),
		CreateTime: timestamppb.New(auth.CreateTime),
		UpdateTime: timestamppb.New(auth.UpdateTime),
	}, nil
}

// Login handles auth login. It verifies the provided credentials, generates a JWT token on success,
// and returns it along with its expiration time. It returns an error if login fails or credentials are incorrect.
func (x *AuthService) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {

	// For login with HTTP GET
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, status.Error(codes.Unknown, "missing metadata")
	// }
	// iden, password, err := extractBasicAuth(md["authorization"])
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

	auth, err := x.store.GetAuthByIdentifier(ctx, in.Identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "incorrect credentials")
		}

		slog.Error("storage get auth by identifier", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(auth.HashedPass), []byte(in.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			slog.Error("bcrypt error mismatch")
			return nil, status.Error(codes.Unauthenticated, "incorrect credentials")
		}
		slog.Error("bcrypt verification failed unexpectedly", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	token, exp, err := createNewToken(auth.ID, pb.Roles(auth.Role))
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
		slog.Error("verify auth token", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
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

// create super admin and demo user
func (x *AuthService) CreateDemoUsers() error {

	ctx := context.TODO()

	adminEmail := os.Getenv("SUPER_ADMIN_EMAIL")
	adminPass := os.Getenv("SUPER_ADMIN_PASS")

	if adminEmail == "" || adminPass == "" {
		return errors.New("some super admin environment variables are not set")
	}

	adminHash, err := hashPassword(adminPass)
	if err != nil {
		return errors.New("hashing super admin password failed")
	}

	_, err = x.store.Create(ctx, &dbNewAuthCredentials{
		Email:       adminEmail,
		HashedPass:  string(adminHash),
		Role:        dbRoles(Roles_SUPER_ADMIN),
		PhoneNumber: nil,
	})
	if err != nil {
		return err
	}

	return nil
}

func (x *AuthService) dispatchCreation(ctx context.Context, role pb.Roles, auth *dbAuthCredentials) error {
	switch role {
	case pb.Roles_CUSTOMER:
		body, err := proto.Marshal(&pb.SyncCustomerCreated{
			CustomerId: auth.ID,
			Email:      auth.Email,
			CreateTime: timestamppb.New(time.Now()),
		})
		if err != nil {
			return err
		}

		if err := x.rabbitmq.publish(ctx, "sync.customer.created", amqp.Publishing{
			Body: body,
		}); err != nil {
			return fmt.Errorf("failed to create customer: %v", err)
		}

	case pb.Roles_RIDER:
		body, err := proto.Marshal(&pb.SyncRiderCreated{
			RiderId:    auth.ID,
			Email:      auth.Email,
			CreateTime: timestamppb.New(time.Now()),
		})
		if err != nil {
			return err
		}

		err = x.rabbitmq.publish(ctx, "sync.rider.created", amqp.Publishing{
			Type: "ihavefood.SyncRiderCreated",
			Body: body,
		})
		if err != nil {
			return fmt.Errorf("failed to create rider: %v", err)
		}

	default:
		return errors.New("invalid role")
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

func extractBasicAuth(authorization []string) (identifier, password string, err error) {

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

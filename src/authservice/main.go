package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/lib/pq"
	pb "github.com/pongsathonn/ihavefood/src/authservice/genproto"
)

// this signing key for testing purpose
var signingKey string

// create singning key when app start
func init() {

	key := make([]byte, 64)

	if _, err := rand.Read(key); err != nil {
		log.Println("generate key failed")
		return
	}

	if len(key) == 0 {
		log.Println("signing key is empty")
		return
	}

	signingKey = string(key)

}

// auth response for authenti iolajskldjaklsjdlasjdl
type auth struct {
	pb.UnimplementedAuthServiceServer

	db *sql.DB
}

func NewAuth(db *sql.DB) *auth {
	return &auth{db: db}
}

func (s *auth) IsValidToken(ctx context.Context, in *pb.IsValidTokenRequest) (*pb.IsValidTokenResponse, error) {

	if valid, err := validateToken(in.Token, []byte(signingKey)); !valid {
		return nil, status.Errorf(codes.Unauthenticated, "token invalid :%v", err)
	}

	return &pb.IsValidTokenResponse{IsValid: true}, nil
}

func (s *auth) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	if in.Username == "" || in.Email == "" || in.Password == "" {
		log.Println("error A: some input empty bro")
		return nil, status.Errorf(codes.InvalidArgument, "failed xxx")
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		log.Println("error B  :", err)
		return nil, status.Errorf(codes.Internal, "failed xxx")
	}

	_, err = s.db.Exec(`INSERT INTO auth_table(username, email, password) VALUES($1, $2, $3)`, in.Username, in.Email, string(hashedPass))

	if err != nil {
		log.Println("error C :", err)
		return nil, status.Errorf(codes.Internal, "failed xxx")
	}

	return &pb.RegisterResponse{}, nil
}

func (s *auth) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {

	// Check if user exists
	var user pb.UserCredentials
	row := s.db.QueryRowContext(ctx, `SELECT username, password FROM auth_table WHERE username=$1`, in.Username)

	if err := row.Scan(&user.Username, &user.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	// Verify password
	correct, err := verifyUserPassword(user.Password, in.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error verifying password: %v", err)
	}
	if !correct {
		return nil, status.Errorf(codes.InvalidArgument, "username or password incorrect")
	}

	// Generate new token and update database
	token, exp, err := createNewToken()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error generating token: %v", err)
	}

	// TODO: Publish event for user login

	return &pb.LoginResponse{AccessToken: token, AccessTokenExp: exp}, nil
}

func verifyUserPassword(hashedPassword, password string) (bool, error) {

	if hashedPassword == "" || password == "" {
		return false, errors.New("invalid input")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false, err
	}
	return true, nil
}

// TODO return error
func validateToken(tokenString string, key []byte) (bool, error) {

	//TODO how its work
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if !token.Valid {
		return false, err
	}

	return true, nil
}

// Symmetric
// Default expire is 5 minutes
func createNewToken() (string, int64, error) {

	// 300 sec = 5 minutes
	addTimeSec := 300
	unixNow := time.Now().Unix()
	expiration := unixNow + int64(addTimeSec)

	claims := &jwt.RegisteredClaims{
		Subject:   "authentication",
		Issuer:    "auth service",
		IssuedAt:  jwt.NewNumericDate(time.Unix(unixNow, 0)),
		ExpiresAt: jwt.NewNumericDate(time.Unix(expiration, 0)),
	}

	// this is token with signing methods and claims
	// i.e Token: &{Header:map[alg:HS256 typ:JWT] Claims:map[exp:1625264265 username:john_doe] Signature:<nil> Raw: Valid:false}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// this is complete signed jwt
	// i.e Signed String: eyJhbGciOiIkpXVCJ9.eyJleHAiOjE9obl9kb2UifQ.4XKl8_uqk9dFekQwQShs
	ss, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", 0, fmt.Errorf("singing failed %v", err)
	}

	return ss, expiration, nil

}

func initPostgres() *sql.DB {

	uri := fmt.Sprintf("postgres://%s:%s@%s:%s/auth_database?sslmode=disable",
		os.Getenv("AUTH_POSTGRES_USER"),
		os.Getenv("AUTH_POSTGRES_PASS"),
		os.Getenv("AUTH_POSTGRES_HOST"),
		os.Getenv("AUTH_POSTGRES_PORT"),
	)

	db, err := sql.Open("postgres", uri)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db

}

func main() {
	db := initPostgres()
	auth := NewAuth(db)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, auth)

	port := os.Getenv("AUTH_SERVER_PORT")
	address := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("auth server starting")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

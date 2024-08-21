package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

type AuthMiddleware interface {
	Authn(next http.Handler) http.Handler
	Authz(next http.Handler) http.Handler
}

type authMiddleware struct{}

// NewMiddleware creates and returns a new Middleware instance with initialized clients.
func NewAuthMiddleware() AuthMiddleware {
	return &authMiddleware{}
}

// authn is authentication middleware
func (m *authMiddleware) Authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, "no authorization in header", http.StatusBadRequest)
			return
		}

		tk := strings.Split(h, " ")
		token := tk[1]

		if valid, err := m.validateToken(token); !valid {
			log.Println(err)
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}

		//TODO

		next.ServeHTTP(w, r)
	})
}

// TODO handler authorization
// authz validate context type might be incorrect fix it
func (m *authMiddleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid Content-Type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateToken checks if the provided token is valid.
func (m *authMiddleware) validateToken(token string) (bool, error) {

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(os.Getenv("AUTH_URI"), opts)
	if err != nil {
		return false, fmt.Errorf("error creating auth client: %v", err)
	}
	client := pb.NewAuthServiceClient(conn)

	req := &pb.IsValidTokenRequest{Token: token}

	resp, err := client.IsValidToken(context.TODO(), req)
	if err != nil {
		log.Println("error validating token:", err)
		return false, err
	}

	if !resp.IsValid {
		log.Println("token is invalid")
		return false, nil
	}

	return true, nil
}

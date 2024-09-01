package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
)

type AuthMiddleware interface {
	Authn(next http.Handler) http.Handler
	Authz(next http.Handler) http.Handler
}

type authMiddleware struct {
	authClient pb.AuthServiceClient
}

// NewMiddleware creates and returns a new Middleware instance with initialized clients.
func NewAuthMiddleware(a pb.AuthServiceClient) AuthMiddleware {
	return &authMiddleware{authClient: a}
}

// Authn is the authentication middleware that verifies the user's token
// and ensures that requests with methods like POST and PUT have the correct Content-Type header.
// It returns an error response if the authentication or validation fails.
func (m *authMiddleware) Authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate the extracted token
		if valid, err := m.validateUserToken(token); !valid {
			log.Printf("Token validation failed: %v\n", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "invalid Content-Type, expected application/json", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Authz is the authorization middleware that restricts access to resources based on user roles.
// This middleware specifically checks if the user has an admin role by validating the token.
// If the token is invalid or the user is not an admin, it returns an unauthorized error.
// Otherwise, it forwards the request to the next handler.
func (m *authMiddleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		valid, err := m.validateAdminToken(token)
		if err != nil {
			log.Printf("Token validation error: %v\n", err)
			http.Error(w, "Token validation failed", http.StatusUnauthorized)
			return
		}

		if !valid {
			http.Error(w, "Access denied: You do not have the required permissions", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validateToken checks if the provided token is valid.
func (m *authMiddleware) validateUserToken(token string) (bool, error) {

	v, err := m.authClient.IsValidToken(context.TODO(), &pb.IsValidTokenRequest{Token: token})
	if err != nil {
		log.Println("error validating token:", err)
		return false, err
	}

	if !v.IsValid {
		log.Println(ErrTokenInvalid)
		return false, nil
	}

	return true, nil
}

func (m *authMiddleware) validateAdminToken(token string) (bool, error) {

	v, err := m.authClient.IsValidAdminToken(context.TODO(), &pb.IsValidAdminTokenRequest{Token: token})
	if err != nil {
		return false, err
	}

	if !v.IsValid {
		log.Println(ErrTokenInvalid)
		return false, nil
	}

	return true, nil
}

// extractToken retrieves and splits the Authorization header, returning the token part.
func extractToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", fmt.Errorf("no authorization in header")
	}

	tk := strings.Split(h, " ")
	if len(tk) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return tk[1], nil
}

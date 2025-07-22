package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

var (
	ErrTokenInvalid = errors.New("token is invalid")
)

type AuthMiddleware struct {
	clientAuth pb.AuthServiceClient
}

func NewAuthMiddleware(a pb.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{clientAuth: a}
}

func (m *AuthMiddleware) Authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			log.Printf("extract token failed : %v\n", err)
			http.Error(w, "failed to extract token", http.StatusBadRequest)
			return
		}

		if valid, err := m.verifyUserToken(token); !valid {
			log.Printf("validate token failed: %v\n", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authz checks user permission to access resource
//
// TODO might implement Role based access control instead checkking only admin token
func (m *AuthMiddleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			log.Printf("extract token failed : %v\n", err)
			http.Error(w, "failed to extract token", http.StatusBadRequest)
			return
		}

		valid, err := m.verifyAdminToken(token)
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

func (m *AuthMiddleware) validateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "invalid Content-Type, expected application/json", http.StatusBadRequest)
				return
			}
		}

		// TODO

		next.ServeHTTP(w, r)
	})
}

// checkUserExists checks whether a user with the given username exists by
// calling the AuthService. returns true if username already exists, otherwise
// false
func (m *AuthMiddleware) checkUsernameExists(username string) (bool, error) {

	res, err := m.clientAuth.CheckUsernameExists(
		context.TODO(),
		&pb.CheckUsernameExistsRequest{Username: username},
	)
	if err != nil {
		return false, err
	}

	return res.Exists, nil
}

// validateToken checks if the provided token is valid by calling the AuthService.
func (m *AuthMiddleware) verifyUserToken(token string) (bool, error) {

	res, err := m.clientAuth.VerifyUserToken(context.TODO(), &pb.VerifyUserTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return false, err
	}

	return res.Valid, nil
}

// verifyAdminToken check if a Token is valid for admin role permission
func (m *AuthMiddleware) verifyAdminToken(token string) (bool, error) {

	res, err := m.clientAuth.VerifyAdminToken(
		context.TODO(),
		&pb.VerifyAdminTokenRequest{AccessToken: token},
	)
	if err != nil {
		return false, err
	}

	return res.Valid, nil
}

// extractToken retrieves and splits the Authorization header, returning the token part.
func extractToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("no authorization in header")
	}

	v := strings.Split(h, " ")
	if len(v) != 2 {
		return "", errors.New("invalid authorization header format")
	}

	return v[1], nil
}

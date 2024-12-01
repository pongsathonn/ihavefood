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

type AuthMiddleware interface {
	// ApplyAuthentication registers routes that require authentication.
	// It applies the `authn` middleware to each of the provided routes.
	//
	// Parameters:
	// - mux: The HTTP multiplexer (ServeMux) to register the routes with.
	// - handler: The HTTP handler that will handle the requests once authentication is successful.
	// - routes: A variadic list of string paths that require authentication.
	//
	// Usage Example:
	//
	//		m.ApplyAuthentication(mux, handler,
	//	        "POST /api/foo",
	//	        "GET /api/bar",
	//		    "/api/profiles",
	//		)
	ApplyAuthentication(mux *http.ServeMux, handler http.Handler, routes ...string)

	ApplyAuthorization(mux *http.ServeMux, handler http.Handler, routes ...string)
}

type authMiddleware struct {
	authClient pb.AuthServiceClient
}

func NewAuthMiddleware(a pb.AuthServiceClient) AuthMiddleware {
	return &authMiddleware{authClient: a}
}

func (m *authMiddleware) ApplyAuthentication(mux *http.ServeMux, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, m.authn(m.validateRequest(handler)))
	}
}

func (m *authMiddleware) ApplyAuthorization(mux *http.ServeMux, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, m.authz(m.validateRequest(handler)))
	}
}

func (m *authMiddleware) authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			log.Printf("extract token failed : %v\n", err)
			http.Error(w, "failed to extract token", http.StatusBadRequest)
			return
		}

		if valid, err := m.validateUserToken(token); !valid {
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
func (m *authMiddleware) authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractToken(r)
		if err != nil {
			log.Printf("extract token failed : %v\n", err)
			http.Error(w, "failed to extract token", http.StatusBadRequest)
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

func (m *authMiddleware) validateRequest(next http.Handler) http.Handler {
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
func (m *authMiddleware) checkUsernameExists(username string) (bool, error) {

	res, err := m.authClient.CheckUsernameExists(
		context.TODO(),
		&pb.CheckUsernameExistsRequest{Username: username},
	)
	if err != nil {
		return false, err
	}

	return res.Exists, nil
}

// validateToken checks if the provided token is valid by calling the AuthService.
func (m *authMiddleware) validateUserToken(token string) (bool, error) {

	res, err := m.authClient.ValidateUserToken(context.TODO(), &pb.ValidateUserTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return false, err
	}

	return res.Valid, nil
}

// validateAdminToken check if a Token is valid for admin role permission
func (m *authMiddleware) validateAdminToken(token string) (bool, error) {

	res, err := m.authClient.ValidateAdminToken(
		context.TODO(),
		&pb.ValidateAdminTokenRequest{AccessToken: token},
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

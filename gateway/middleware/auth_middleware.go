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
	ApplyAuthentication(mux *http.ServeMux, handler http.Handler, routes ...string)
	ApplyAuthorization(mux *http.ServeMux, handler http.Handler, routes ...string)
	ApplyFullAuth(mux *http.ServeMux, handler http.Handler, routes ...string)
}

type authMiddleware struct {
	authClient pb.AuthServiceClient
}

func NewAuthMiddleware(a pb.AuthServiceClient) AuthMiddleware {
	return &authMiddleware{authClient: a}
}

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
func (m *authMiddleware) ApplyAuthentication(mux *http.ServeMux, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, m.authn(handler))
	}
}

// ApplyAuthorization registers routes that require authorization.
// It applies the `authz` middleware to each of the provided routes.
//
// See example with ApplyAuthentication
func (m *authMiddleware) ApplyAuthorization(mux *http.ServeMux, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, m.authz(handler))
	}
}

// ApplyFullAuth registers routes that require both authentication and authorization.
// It applies the `authn` middleware first and then the `authz` middleware to each route.
//
// See example with ApplyAuthentication
func (m *authMiddleware) ApplyFullAuth(mux *http.ServeMux, handler http.Handler, routes ...string) {
	for _, route := range routes {
		mux.Handle(route, m.authn(m.authz(handler)))
	}
}

// authn is the authentication middleware that verifies the user's token
// and ensures that requests with methods like POST and PUT have the correct Content-Type header.
// It returns an error response if the authentication or validation fails.
func (m *authMiddleware) authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractAuth(r)
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

// authz is the authorization middleware that restricts access to resources based on user roles.
// This middleware specifically checks if the user has an admin role by validating the token.
// If the token is invalid or the user is not an admin, it returns an unauthorized error.
// Otherwise, it forwards the request to the next handler.
func (m *authMiddleware) authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := extractAuth(r)
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

// extractAuth retrieves and splits the Authorization header, returning the token part.
func extractAuth(r *http.Request) (string, error) {
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

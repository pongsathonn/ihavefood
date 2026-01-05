package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"strings"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/pongsathonn/ihavefood/api-gateway/genproto"
)

var (
	signingKey []byte

	ErrTokenInvalid = errors.New("token is invalid")
)

type GatewayClaims struct {
	Role pb.Roles `json:"role"`
	jwt.RegisteredClaims
}

func LoadSigningkey() {
	key := os.Getenv("JWT_SIGNING_KEY")
	if key == "" {
		log.Fatal("missing JWT_SIGNING_KEY environment variable")
	}
	signingKey = []byte(key)
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenStr, err := extractToken(r)
		if err != nil {
			log.Printf("extract token failed : %v\n", err)
			http.Error(w, "failed to extract token", http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, new(GatewayClaims), func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signingKey, nil
		})
		if err != nil {
			log.Printf("validate token failed: %v\n", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// check permission for resource under /admin
		if strings.HasPrefix(r.URL.Path, "/api/admin/") {
			claims, ok := token.Claims.(*GatewayClaims)
			if !ok || claims.Role != pb.Roles_ROLES_ADMIN {
				http.Error(w, "Access denied: You do not have the required permissions", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
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

// func validateRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
// 		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
// 			if r.Header.Get("Content-Type") != "application/json" {
// 				http.Error(w, "invalid Content-Type, expected application/json", http.StatusBadRequest)
// 				return
// 			}
// 		}
//
// 		next.ServeHTTP(w, r)
// 	})
// }

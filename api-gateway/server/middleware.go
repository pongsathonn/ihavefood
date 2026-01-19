package server

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
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

		cookie, err := r.Cookie("access-token")
		if err != nil {
			slog.Error("unable to read cookie", "err", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, new(GatewayClaims), func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
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

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

// TODO find good name ja
type XXX interface {
	Authn(next http.Handler) http.Handler
	Authz(next http.Handler) http.Handler
	VerifyPlaceOrder(next http.Handler) http.Handler
}

// Middleware holds gRPC clients for interacting with various microservices.
type middleware struct {
}

// NewMiddleware creates and returns a new Middleware instance with initialized clients.
// TODO setup grpc client from main
func NewMiddleware() XXX {

	return &middleware{}
}

// authn is authentication middleware
func (m *middleware) Authn(next http.Handler) http.Handler {
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
func (m *middleware) Authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid Content-Type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// verifyPlaceOrder verifies the place order request by checking menu availability.
func (m *middleware) VerifyPlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		// re-assign body to body request
		r.Body = io.NopCloser(bytes.NewReader(body))

		var req pb.CheckAvailableMenuRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal: %v", err), http.StatusBadRequest)
			return
		}

		if req.RestaurantName == "" || len(req.Menus) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		availMenu, err := m.availableMenu(req.RestaurantName, req.Menus)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to verify: %v", err), http.StatusBadRequest)
			return
		}

		if !availMenu /*|| !availCoupon*/ {
			http.Error(w, "menu not available", http.StatusBadRequest)
			return
		}

		//TODO verify valid coupon

		next.ServeHTTP(w, r)
	})
}

// availableMenu checks if the menu items for a given restaurant are available.
func (m *middleware) availableMenu(restauName string, menus []*pb.Menu) (bool, error) {

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(os.Getenv("RESTAURANT_URI"), opts)
	if err != nil {
		return false, fmt.Errorf("error creating restaurant client: %v", err)
	}
	client := pb.NewRestaurantServiceClient(conn)

	req := &pb.CheckAvailableMenuRequest{
		RestaurantName: restauName,
		Menus:          menus,
	}

	check, err := client.CheckAvailableMenu(context.TODO(), req)
	if err != nil {
		return false, err
	}

	// 0: available, 1: unavailable, 2: unknown
	checkNumber := check.Available.Number()
	if checkNumber != 0 {
		return false, fmt.Errorf("menu status not available with number: %d", checkNumber)
	}

	return true, nil
}

// validateToken checks if the provided token is valid.
func (m *middleware) validateToken(token string) (bool, error) {

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

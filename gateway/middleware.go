package main

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

func authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, "no authorization in header", http.StatusBadRequest)
			return
		}

		tk := strings.Split(h, " ")
		token := tk[1]

		if valid, err := validateToken(token); !valid {
			log.Println(err)
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}

		//TODO

		next.ServeHTTP(w, r)

	})
}

func authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid Content-Type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func validateToken(token string) (bool, error) {

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv("AUTH_URI"), opts)
	if err != nil {
		log.Println("error new auth client :", err)
		return false, err
	}

	client := pb.NewAuthServiceClient(conn)

	resp, err := client.IsValidToken(context.TODO(), &pb.IsValidTokenRequest{Token: token})
	if err != nil {
		log.Println("error not valid :", err)
		return false, err
	}

	if !resp.IsValid {
		log.Println("token invalid ja ")
		return false, err
	}

	return true, nil

}

// TODO improve error handler and refactor code
func verifyPlaceOrder(next http.Handler) http.Handler {
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

		availMenu, err := availableMenu(req.RestaurantName, req.Menus)
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

func availableMenu(restauName string, menus []*pb.Menu) (bool, error) {

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(os.Getenv("RESTAURANT_URI"), opts)
	if err != nil {
		log.Println("error create restau client in gateway", err)
		return false, err
	}
	client := pb.NewRestaurantServiceClient(conn)

	check, err := client.CheckAvailableMenu(context.TODO(),
		&pb.CheckAvailableMenuRequest{
			RestaurantName: restauName,
			Menus:          menus,
		})
	if err != nil {
		log.Println("AB@", err)
		return false, err
	}

	/* 0 avail, 1 unavail, 2 uknown */
	cn := check.Available.Number()
	if cn != 0 {
		return false, fmt.Errorf("menu status %d", cn)
	}

	return true, nil
}

func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

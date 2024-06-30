package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pongsathonn/food-delivery/gateway/genproto"
)

// authn ( authentication ) who are you
// check token , user credentials correct blabla
func authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, "no aothorization in header", http.StatusBadRequest)
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

// authz ( authorization  ) what are you allowed to do ?
// check user permission for access resource
// i.e User allowed to Delete this ?
func authz(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid Content-Type", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)

	})
}

// Call authService to validate token
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

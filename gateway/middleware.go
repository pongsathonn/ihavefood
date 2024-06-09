package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/rand"

	"github.com/golang-jwt/jwt/v5"
)

func validatePlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})
}

func authn(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("Content-Type") != "application/json" {
			log.Println("invalid Content-Type")
			return
		}

		h := r.Header.Get("Authorization")
		tk := strings.Split(h, " ")

		var token string
		token = tk[1]

		if token != "2222" {
			fmt.Println(token)
			return
		}

		sk := os.Getenv("SIGNING_KEY")

		if sk == "" {
			log.Println("singing_key not found or empty")
			return
		}

		if err := validateToken(token, []byte(sk)); err != nil {
			log.Println("unauthorized", err)
			return
		}

		next.ServeHTTP(w, r)

	})
}

// TODO return error
func validateToken(tokenString string, key []byte) error {

	//TODO how its work
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if !token.Valid {
		return err
	}

	return nil
}

// Symmetric
func createNewToken() (string, error) {

	sk := os.Getenv("SIGNING_KEY")
	if sk == "" {
		return "", fmt.Errorf("signing key not found or empty")
	}

	// 300 sec is 5 minutes
	unixNow := time.Now().Unix()
	unixFiveMinutes := unixNow + 300

	claims := &jwt.RegisteredClaims{
		Subject:   "authentication",
		Issuer:    "lineman wongnai team",
		IssuedAt:  jwt.NewNumericDate(time.Unix(unixNow, 0)),
		ExpiresAt: jwt.NewNumericDate(time.Unix(unixFiveMinutes, 0)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// this ss from signedString used as share secret key
	ss, err := token.SignedString([]byte(sk))
	if err != nil {
		return "", fmt.Errorf("singing failed %v", err)
	}

	return ss, nil

}

// Regenerate everytime when starting server
func generateKey() {

	key := os.Getenv("SIGNING_KEY")

	if key == "" {
		key := make([]byte, 64)

		_, err := rand.Read(key)
		if err != nil {
			log.Println("generate key failed")
			return
		}

		os.Setenv("SIGNING_KEY", string(key))

		ss := os.Getenv("SIGNING_KEY")
		if ss != "" {
			log.Println("new signing key is generated")
		}
	}

}

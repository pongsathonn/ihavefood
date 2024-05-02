package main

import (
	"fmt"
	"log"
	"net/http"
)

// TODO: gRPC gateway

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/food", OrderFood)

	log.Println("starting server")
	log.Println(http.ListenAndServe("localhost:1150", mux))

}

func OrderFood(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello Client!")
}

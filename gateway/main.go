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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/pongsathonn/food-delivery/gateway/genproto"
)

func main() {

	/*
		//generate jwt key
		//FIXME sometime this not working
		if err := generateKey(); err != nil {
			log.Println(err)
		}
	*/

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:    "  ",
					Multiline: true, // Optional, implied by presence of "Indent".
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			}),
	)

	//TODO handle error from register
	pb.RegisterUserServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("USER_URI"), opts)
	pb.RegisterCouponServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("COUPON_URI"), opts)
	pb.RegisterOrderServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("ORDER_URI"), opts)
	pb.RegisterRestaurantServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("RESTAURANT_URI"), opts)

	mux := http.NewServeMux()
	mux.Handle("POST /api/orders/place-order", verifyPlaceOrder(gwmux))
	mux.Handle("/", gwmux)
	log.Println("gateway starting")

	gwp := os.Getenv("GATEWAY_PORT")
	log.Println("GATEWAY PORT =", gwp)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", gwp), prettierJSON(mux)))
}

// TODO improve error handler and refactor code
func verifyPlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
			return
		}
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

		// can do post logic here
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

// json response pretty
func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

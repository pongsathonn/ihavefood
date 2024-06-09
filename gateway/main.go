package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/pongsathonn/food-delivery/gateway/genproto"
)

func main() {

	//generate jwt key
	generateKey()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	gwmux := runtime.NewServeMux()

	pb.RegisterUserServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("USER_URI"), opts)
	pb.RegisterCouponServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("COUPON_URI"), opts)
	pb.RegisterOrderServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("ORDER_URI"), opts)
	pb.RegisterRestaurantServiceHandlerFromEndpoint(context.TODO(), gwmux, os.Getenv("RESTAURANT_URI"), opts)

	mux := http.NewServeMux()

	//TODO validate user input, rate limit
	mux.Handle("/", gwmux)
	mux.Handle("/register", gwmux)
	//mux.Handle("/api/*", authn(gwmux))
	mux.Handle("/api", gwmux)

	mux.Handle("/api/orders/place-order", verifyPlaceOrder(gwmux))

	log.Println("gateway starting")
	// FIXME: ListenAndServeTLS
	log.Fatal(http.ListenAndServe(os.Getenv("GATEWAY_URI"), mux))
}

func verifyPlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//TODO validate input

		conn, err := grpc.NewClient(os.Getenv("RESTAURANT_URI"))
		if err != nil {
			log.Fatal(err)
		}

		rc := pb.NewRestaurantServiceClient(conn)
		cc := pb.NewCouponServiceClient(conn)

		ava, err := rc.IsAvaliableMenu("krapao")
		if err != nil {
			log.Println(err)
			return
		}

		if !ava {
			log.Println("menus not avaliable for order")
			return
		}

		couponTest := "cpd10912"
		valid, err := cc.verifyCoupon(couponTest)
		if err != nil {
			log.Println(err)
			return
		}

		if !valid {
			log.Println("coupon not valid")
			return
		}

		next.ServeHTTP(w, r)

		log.Println("can do some logic after here")
	})
}

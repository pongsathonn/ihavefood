package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/your-username/your-project/genproto"
)

func main() {

	// above this part is gateway

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				http.Error(w, "No Content", http.StatusNoContent)
				return
			}

			mux.ServeHTTP(w, r)
		})
		corsHandler.ServeHTTP(w, r)
	})

	log.Println("Gateway server starting on port :2020")
	s := ":2020"
	log.Fatal(http.ListenAndServe(s, handler))
}

func gateway() {
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:    "  ",
					Multiline: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			}),
		runtime.WithForwardResponseOption(setStatus),
	)

	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	services := []struct {
		registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
		endpoint     string
	}{
		{pb.RegisterUserServiceHandlerFromEndpoint, "USER_URI"},
		{pb.RegisterCouponServiceHandlerFromEndpoint, "COUPON_URI"},
		{pb.RegisterOrderServiceHandlerFromEndpoint, "ORDER_URI"},
		{pb.RegisterRestaurantServiceHandlerFromEndpoint, "RESTAURANT_URI"},
		{pb.RegisterDeliveryServiceHandlerFromEndpoint, "DELIVERY_URI"},
		{pb.RegisterAuthServiceHandlerFromEndpoint, "AUTH_URI"},
	}

	for _, handler := range services {
		if err := handler.registerFunc(ctx, gwmux, os.Getenv(handler.endpoint), opts); err != nil {
			log.Fatalf("failed to register service handler: %v", err)
		}
	}

	auth, service := gateway.InitMiddleware()

	mux := http.NewServeMux()
	mux.Handle("POST /auth/login", gwmux)
	mux.Handle("POST /auth/register", gwmux)
	mux.Handle("PUT /auth/users/roles", auth.Authz(gwmux))
	mux.Handle("POST /api/orders/place-order", auth.Authn(service.VerifyPlaceOrder(gwmux)))
	mux.Handle("POST /api/users", auth.Authn(gwmux))
	mux.Handle("DELETE /api/*", auth.Authn(auth.Authz(gwmux)))
	mux.Handle("/api/*", auth.Authn(gwmux))
}

func server() {
}

// setStatus handles specific response types
func setStatus(ctx context.Context, w http.ResponseWriter, m protoreflect.ProtoMessage) error {
	switch m.(type) {
	case *pb.RegisterResponse:
		w.WriteHeader(http.StatusCreated)
	}
	return nil
}

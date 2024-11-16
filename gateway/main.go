package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pongsathonn/ihavefood/gateway/middleware"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

const ihavefood = `
=======================================================================
██╗██╗  ██╗ █████╗ ██╗   ██╗███████╗███████╗ ██████╗  ██████╗ ██████╗
██║██║  ██║██╔══██╗██║   ██║██╔════╝██╔════╝██╔═══██╗██╔═══██╗██╔══██╗
██║███████║███████║██║   ██║█████╗  █████╗  ██║   ██║██║   ██║██║  ██║
██║██╔══██║██╔══██║╚██╗ ██╔╝██╔══╝  ██╔══╝  ██║   ██║██║   ██║██║  ██║
██║██║  ██║██║  ██║ ╚████╔╝ ███████╗██║     ╚██████╔╝╚██████╔╝██████╔╝
╚═╝╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝      ╚═════╝  ╚═════╝ ╚═════╝
=======================================================================
`

func main() {

	gwmux := newGatewaymux()

	if err := registerServiceHandlers(context.TODO(), gwmux); err != nil {
		log.Fatalf("failed to register service handler: %v", err)
	}
	auth, err := initAuthMiddleware()
	if err != nil {
		log.Fatalf("failed to initialize auth middleware: %v", err)
	}

	mux := setupHTTPMuxAndAuth(auth, gwmux)
	handler := prettierJSON(cors(mux))

	log.Println("+++++ GATEWAY STARTING +++++")
	log.Println(ihavefood)

	s := fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT"))
	log.Fatal(http.ListenAndServe(s, handler))

}

func newGatewaymux() *runtime.ServeMux {
	return runtime.NewServeMux(
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
		runtime.WithForwardResponseOption(setStatus),
	)

}

// setStatus handle specific response types
func setStatus(ctx context.Context, w http.ResponseWriter, m protoreflect.ProtoMessage) error {

	switch m.(type) {
	case *pb.RegisterResponse:
		w.WriteHeader(http.StatusCreated)
	}
	// keep default behavior
	return nil
}

// Register service handlers
func registerServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux) error {

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
			return err
		}
	}

	return nil
}

func initAuthMiddleware() (middleware.AuthMiddleware, error) {

	conn, err := grpc.Dial(
		os.Getenv("AUTH_URI"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return middleware.NewAuthMiddleware(pb.NewAuthServiceClient(conn)), nil
}

func setupHTTPMuxAndAuth(m middleware.AuthMiddleware, gwmux *runtime.ServeMux) http.Handler {
	mux := http.NewServeMux()
	m.ApplyFullAuth(mux, gwmux,
		"DELETE /api/*",
	)
	m.ApplyAuthorization(mux, gwmux,
		"/auth/users/roles",
	)
	m.ApplyAuthentication(mux, gwmux,
		"/api/orders/place-order",
		"/api/users",
		"/api/*",
	)
	mux.Handle("/", gwmux)
	return mux
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//    For production
		// 	  swg := fmt.Sprintf("http://localhost:%s", os.Getenv("SWAGGER_UI_PORT"))
		// 	  w.Header().Add("Access-Control-Allow-Origin", swg)

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

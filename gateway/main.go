package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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

	slog.Info("Gateway initialization starting")
	log.Print(ihavefood)
	if err := registerServiceHandlers(context.TODO(), gwmux); err != nil {
		log.Fatalf("failed to register service handler: %v", err)
	}

	auth, err := initAuthMiddleware()
	if err != nil {
		log.Fatalf("failed to initialize auth middleware: %v", err)
	}

	mux := setupHTTPMuxAndAuth(auth, gwmux)
	handler := prettierJSON(cors(mux))

	gwport := os.Getenv("GATEWAY_PORT")
	addr := fmt.Sprintf(":%s", gwport)

	slog.Info(fmt.Sprintf("Gateway listening on port :%s", gwport))

	log.Fatal(http.ListenAndServe(addr, handler))

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

// setStatus sets response code for specific type.
//
// the default behavior success code for gRPC-gateway is 200.
// if you want to modify such as change created response from 200 to 201
// then add case in setStatus code
func setStatus(ctx context.Context, w http.ResponseWriter, m protoreflect.ProtoMessage) error {

	/*
		switch m.(type) {
		case *pb.RegisterUserResponse:
			w.WriteHeader(http.StatusCreated)
		}
	*/

	// keep default behavior
	return nil
}

// Register service handlers
func registerServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux) error {

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	services := []struct {
		name         string
		registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
		endpoint     string
	}{
		{"ProfileService", pb.RegisterProfileServiceHandlerFromEndpoint, "PROFILE_URI"},
		{"CouponService", pb.RegisterCouponServiceHandlerFromEndpoint, "COUPON_URI"},
		{"OrderService", pb.RegisterOrderServiceHandlerFromEndpoint, "ORDER_URI"},
		{"RestaurantService", pb.RegisterRestaurantServiceHandlerFromEndpoint, "RESTAURANT_URI"},
		{"DeliveryService", pb.RegisterDeliveryServiceHandlerFromEndpoint, "DELIVERY_URI"},
		{"AuthService", pb.RegisterAuthServiceHandlerFromEndpoint, "AUTH_URI"},
	}

	for _, handler := range services {
		addr := os.Getenv(handler.endpoint)
		if err := handler.registerFunc(ctx, gwmux, addr, opts); err != nil {
			return err
		}
	}

	slog.Info("Register every services successfully")
	return nil
}

func initAuthMiddleware() (middleware.AuthMiddleware, error) {

	addr := os.Getenv("AUTH_URI")
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	slog.Info("Auth middleware initialized successfully", "address", addr)

	return middleware.NewAuthMiddleware(pb.NewAuthServiceClient(conn)), nil
}

func setupHTTPMuxAndAuth(m middleware.AuthMiddleware, gwmux *runtime.ServeMux) http.Handler {
	mux := http.NewServeMux()

	// update role and delete need "admin" role.
	m.ApplyAuthorization(mux, gwmux,
		"/auth/users/roles",
		"DELETE /api/*",
	)

	m.ApplyAuthentication(mux, gwmux,
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
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
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

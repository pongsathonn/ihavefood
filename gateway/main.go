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
	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
	"github.com/pongsathonn/ihavefood/gateway/middleware"
)

func main() {
	gwmux := newGatewaymux()
	ctx := context.Background()
	if err := registerServiceHandlers(ctx, gwmux); err != nil {
		log.Fatalf("failed to register service handler: %v", err)
	}

	auth, service := initMiddleware()
	mux := setupHTTPMux(auth, service, gwmux)
	handler := prettierJSON(cors(mux))

	log.Println("gateway starting")
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

func initMiddleware() (middleware.AuthMiddleware, middleware.ServiceMiddleware) {

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	authConn, err := grpc.Dial(os.Getenv("AUTH_URI"), opts)
	if err != nil {
		log.Fatalf("error creating restaurant client: %v", err)
	}
	authClient := pb.NewAuthServiceClient(authConn)

	restaurantConn, err := grpc.Dial(os.Getenv("RESTAURANT_URI"), opts)
	if err != nil {
		log.Fatalf("error creating restaurant client: %v", err)
	}
	restaurantClient := pb.NewRestaurantServiceClient(restaurantConn)

	couponConn, err := grpc.Dial(os.Getenv("COUPON_URI"), opts)
	if err != nil {
		log.Fatalf("error creating coupon client: %v", err)
	}
	couponClient := pb.NewCouponServiceClient(couponConn)

	orderConn, err := grpc.Dial(os.Getenv("ORDER_URI"), opts)
	if err != nil {
		log.Fatalf("error creating order client: %v", err)
	}
	orderClient := pb.NewOrderServiceClient(orderConn)

	deliveryConn, err := grpc.Dial(os.Getenv("DELIVERY_URI"), opts)
	if err != nil {
		log.Fatalf("error creating delivery client: %v", err)
	}
	deliveryClient := pb.NewDeliveryServiceClient(deliveryConn)

	userConn, err := grpc.Dial(os.Getenv("USER_URI"), opts)
	if err != nil {
		log.Fatalf("error creating user client: %v", err)
	}
	userClient := pb.NewUserServiceClient(userConn)

	// Service Middleware Configuration
	cfg := middleware.ServiceMiddlewareConfig{
		RestaurantClient: restaurantClient,
		CouponClient:     couponClient,
		OrderClient:      orderClient,
		DeliveryClient:   deliveryClient,
		UserClient:       userClient,
	}

	auth := middleware.NewAuthMiddleware(authClient)
	service := middleware.NewServiceMiddleware(cfg)

	return auth, service
}

func setupHTTPMux(auth middleware.AuthMiddleware, svc middleware.ServiceMiddleware, gwmux *runtime.ServeMux) http.Handler {

	var (
		authz = auth.Authz
		authn = auth.Authn
	)

	mux := http.NewServeMux()

	mux.Handle("POST /auth/login", gwmux)
	mux.Handle("POST /auth/register", gwmux)
	mux.Handle("PUT /auth/users/roles", authz(gwmux))

	mux.Handle("POST /api/orders/place-order", authn(svc.VerifyPlaceOrder(gwmux)))
	mux.Handle("POST /api/users", authn(gwmux))

	mux.Handle("DELETE /api/*", authn(authz(gwmux)))
	mux.Handle("/api/*", authn(gwmux))

	return mux
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

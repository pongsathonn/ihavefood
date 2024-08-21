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

	authMW, serviceMW, err := initMiddleware()
	if err != nil {
		log.Fatalf("failed to init middleware: %v", err)
	}
	mux := setupHTTPMux(authMW, serviceMW, gwmux)

	log.Println("gateway starting")

	s := fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT"))
	log.Fatal(http.ListenAndServe(s, mux))
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
	)

}

// Register service handlers
func registerServiceHandlers(ctx context.Context, gwmux *runtime.ServeMux) error {

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	services := []struct {
		regsiterFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
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
		if err := handler.regsiterFunc(ctx, gwmux, os.Getenv(handler.endpoint), opts); err != nil {
			return err
		}
	}

	return nil
}

// TODO try to understand this
func initMiddleware() (middleware.AuthMiddleware, middleware.ServiceMiddleware, error) {

	clientConfig := map[string]func(*grpc.ClientConn) interface{}{
		"RESTAURANT_URI": func(conn *grpc.ClientConn) interface{} { return pb.NewRestaurantServiceClient(conn) },
		"COUPON_URI":     func(conn *grpc.ClientConn) interface{} { return pb.NewCouponServiceClient(conn) },
		"ORDER_URI":      func(conn *grpc.ClientConn) interface{} { return pb.NewOrderServiceClient(conn) },
		"DELIVERY_URI":   func(conn *grpc.ClientConn) interface{} { return pb.NewDeliveryServiceClient(conn) },
		"USER_URI":       func(conn *grpc.ClientConn) interface{} { return pb.NewUserServiceClient(conn) },
	}

	clients := make(map[string]interface{})
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	for key, createClient := range clientConfig {
		conn, err := grpc.Dial(os.Getenv(key), opts)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating %s client: %v", key, err)
		}
		clients[key] = createClient(conn)
	}

	cfg := middleware.ServiceMiddlewareConfig{
		RestaurantClient: clients["RESTAURANT_URI"].(pb.RestaurantServiceClient),
		CouponClient:     clients["COUPON_URI"].(pb.CouponServiceClient),
		OrderClient:      clients["ORDER_URI"].(pb.OrderServiceClient),
		DeliveryClient:   clients["DELIVERY_URI"].(pb.DeliveryServiceClient),
		UserClient:       clients["USER_URI"].(pb.UserServiceClient),
	}

	serviceMW := middleware.NewServiceMiddleware(cfg)
	authMW := middleware.NewAuthMiddleware()

	return authMW, serviceMW, nil
}

func setupHTTPMux(auth middleware.AuthMiddleware, svc middleware.ServiceMiddleware, gwmux *runtime.ServeMux) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("/auth/login", gwmux)
	mux.Handle("/auth/register", gwmux)
	mux.Handle("DELETE /api/*", auth.Authz(gwmux))

	// production use this
	mux.Handle("POST /api/orders/place-order", auth.Authn(svc.VerifyPlaceOrder(gwmux)))
	mux.Handle("POST /api/users", auth.Authn(gwmux))

	mux.Handle("/api/*", auth.Authn(gwmux))
	mux.Handle("/", gwmux)

	return prettierJSON(mux)
}

func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

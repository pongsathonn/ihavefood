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

	log.Println("gateway starting")

	mw := middleware.NewMiddleware()
	mux := setupHTTPMux(mw, gwmux)

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

func setupHTTPMux(mw middleware.XXX, gwmux *runtime.ServeMux) http.Handler {

	mux := http.NewServeMux()
	mux.Handle("/auth/login", gwmux)
	mux.Handle("/auth/register", gwmux)
	mux.Handle("DELETE /api/*", mw.Authz(gwmux))

	// production use this
	mux.Handle("POST /api/orders/place-order", mw.Authn(mw.VerifyPlaceOrder(gwmux)))
	mux.Handle("POST /api/users", mw.Authn(gwmux))

	mux.Handle("/api/*", mw.Authn(gwmux))
	mux.Handle("/", gwmux)

	return prettierJSON(mux)
}

func prettierJSON(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "application/json+pretty")
		h.ServeHTTP(w, r)
	})
}

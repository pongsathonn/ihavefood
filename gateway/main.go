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
)

func main() {

	gwmux := newGatewaymux()

	ctx := context.Background()
	if err := registerServiceHandlers(ctx, gwmux); err != nil {
		log.Fatalf("failed to register service handler: %v", err)
	}

	log.Println("gateway starting")

	mux := setupHTTPMux(gwmux)
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

func setupHTTPMux(gwmux *runtime.ServeMux) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", gwmux)
	//mux.Handle("DELETE /api/*", authz(gwmux))

	// production use this
	mux.Handle("POST /api/orders/place-order", authn(verifyPlaceOrder(gwmux)))
	mux.Handle("/api/*", authn(gwmux))

	// testing
	//mux.Handle("/api/*", (gwmux))
	//mux.Handle("POST /api/orders/place-order", (verifyPlaceOrder(gwmux)))

	mux.Handle("/", gwmux)

	return prettierJSON(mux)
}

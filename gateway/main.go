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
	fmt.Print(ihavefood)

	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:    "  ",
			Multiline: true, // Optional, implied by presence of "Indent".
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", marshaler),
	)
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	registerService(context.TODO(), gwmux, []grpc.DialOption{opt})
	conn, err := grpc.Dial(os.Getenv("AUTH_URI"), opt)
	if err != nil {
		log.Fatal(err)
	}
	auth := middleware.NewAuthMiddleware(pb.NewAuthServiceClient(conn))

	// Update role and DELETE methods requires "admin" role.
	http.Handle("PATCH /auth/users/roles", auth.Authz(gwmux))
	http.Handle("DELETE /api/*", auth.Authz(gwmux))
	http.Handle("/api/users/*", auth.Authn(gwmux))

	http.Handle("/api/*", auth.Authn(gwmux))
	http.Handle("/", gwmux)

	gwport := os.Getenv("GATEWAY_PORT")
	slog.Info(fmt.Sprintf("Gateway listening on port :%s", gwport))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", gwport), prettierJSON(cors(gwmux))); err != nil {
		log.Fatalf("Failed to Serve %v", err)
	}

}

func registerService(ctx context.Context, gwmux *runtime.ServeMux, opts []grpc.DialOption) {
	services := map[string]func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		"PROFILE_URI":    pb.RegisterProfileServiceHandlerFromEndpoint,
		"COUPON_URI":     pb.RegisterCouponServiceHandlerFromEndpoint,
		"ORDER_URI":      pb.RegisterOrderServiceHandlerFromEndpoint,
		"RESTAURANT_URI": pb.RegisterRestaurantServiceHandlerFromEndpoint,
		"DELIVERY_URI":   pb.RegisterDeliveryServiceHandlerFromEndpoint,
		"AUTH_URI":       pb.RegisterAuthServiceHandlerFromEndpoint,
	}

	for envVar, registerFunc := range services {
		uri := os.Getenv(envVar)
		if uri == "" {
			log.Fatalf("%s is not set", envVar)
		}
		if err := registerFunc(ctx, gwmux, uri, opts); err != nil {
			log.Fatalf("Failed to register %s: %v", envVar, err)
		}
	}
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

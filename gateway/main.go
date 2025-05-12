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
<<<<<<< HEAD

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", marshaler),
	)
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	registerService(context.TODO(), gwmux, []grpc.DialOption{opt})
=======
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", marshaler),
	)

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	opts := []grpc.DialOption{opt}

	ctx := context.TODO()
	err := pb.RegisterProfileServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("PROFILE_URI"), opts)
	err = pb.RegisterCouponServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("COUPON_URI"), opts)
	err = pb.RegisterOrderServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("ORDER_URI"), opts)
	err = pb.RegisterRestaurantServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("RESTAURANT_URI"), opts)
	err = pb.RegisterDeliveryServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("DELIVERY_URI"), opts)
	err = pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gwmux, os.Getenv("AUTH_URI"), opts)
>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
	conn, err := grpc.Dial(os.Getenv("AUTH_URI"), opt)
	if err != nil {
		log.Fatal(err)
	}
<<<<<<< HEAD
	auth := middleware.NewAuthMiddleware(pb.NewAuthServiceClient(conn))

	// Update role and DELETE methods requires "admin" role.
=======

	auth := middleware.NewAuthMiddleware(pb.NewAuthServiceClient(conn))

	// Update role and DELETE methods need "admin" role.

>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
	http.Handle("PATCH /auth/users/roles", auth.Authz(gwmux))
	http.Handle("DELETE /api/*", auth.Authz(gwmux))
	http.Handle("/api/users/*", auth.Authn(gwmux))

	http.Handle("/api/*", auth.Authn(gwmux))
	http.Handle("/", gwmux)

<<<<<<< HEAD
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

=======
	handler := prettierJSON(cors(gwmux))

	gwport := os.Getenv("GATEWAY_PORT")
	addr := fmt.Sprintf(":%s", gwport)

	slog.Info(fmt.Sprintf("Gateway listening on port :%s", gwport))
	log.Fatal(http.ListenAndServe(addr, handler))

}

>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
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

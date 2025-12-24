package server

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net/http"
	"os"

	"context"
	"net/url"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/credentials/oauth"

	"log"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/pongsathonn/ihavefood/api-gateway/genproto"
)

func cors(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// prefight request check with OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
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

func newGateway() http.Handler {

	mars := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:    "  ",
			Multiline: true, // Optional, implied by presence of "Indent".
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", mars),
		// runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(cl)),
	)

	for env, f := range map[string]func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		"CUSTOMER_URI": pb.RegisterCustomerServiceHandlerFromEndpoint,
		"COUPON_URI":   pb.RegisterCouponServiceHandlerFromEndpoint,
		"ORDER_URI":    pb.RegisterOrderServiceHandlerFromEndpoint,
		"MERCHANT_URI": pb.RegisterMerchantServiceHandlerFromEndpoint,
		"DELIVERY_URI": pb.RegisterDeliveryServiceHandlerFromEndpoint,
		"AUTH_URI":     pb.RegisterAuthServiceHandlerFromEndpoint,
	} {

		ctx := context.Background()
		uri := os.Getenv(env)
		if uri == "" {
			log.Fatalf("%s is not set", env)
		}

		// NOTE: audience must include scheme
		tokenSource, err := idtoken.NewTokenSource(ctx, uri)
		if err != nil {
			log.Fatalf("idtoken.NewTokenSource failed: %v", err)
		}

		parsedURL, err := url.Parse(uri)
		if err != nil {
			log.Fatalf("Failed to parse uri: %v", err)
		}

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
			grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: tokenSource}),
		}

		host := parsedURL.Host
		if parsedURL.Port() == "" {
			host = host + ":443"
		}

		conn, err := grpc.NewClient(host, opts...)
		if err != nil {
			log.Fatalf("Failed to dial %s for health check: %v", env, err)
		}

		healthClient := healthgrpc.NewHealthClient(conn)
		res, err := healthClient.Check(ctx, &healthgrpc.HealthCheckRequest{})
		if err != nil || res.GetStatus() != healthgrpc.HealthCheckResponse_SERVING {
			log.Fatalf("Health check failed for %s: %v", uri, err)
		}

		if err := f(ctx, mux, host, opts); err != nil {
			log.Fatalf("Failed to register %s: %v", env, err)
		}
	}

	return mux
}

func initAuthMiddleware() (*AuthMiddleware, error) {
	ctx := context.Background()

	authURI := os.Getenv("AUTH_URI")
	tokenSource, err := idtoken.NewTokenSource(ctx, authURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to create token source: %v", err)
	}

	parsedURL, err := url.Parse(authURI)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse AUTH_URI: %v", err)
	}

	authConn, err := grpc.NewClient(parsedURL.Host,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: tokenSource}),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create auth client: %v", err)

	}
	return NewAuthMiddleware(pb.NewAuthServiceClient(authConn)), nil
}

func Run() error {

	gwmux := newGateway()

	auth, err := initAuthMiddleware()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/merchants", gwmux)
	mux.Handle("DELETE /api/", auth.Authz(gwmux))
	mux.Handle("/api/", auth.Authn(gwmux))
	mux.Handle("/", gwmux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &http.Server{
		Addr:    ":" + port,
		Handler: prettierJSON(cors(mux)),
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

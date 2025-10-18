package gateway

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	pb "github.com/pongsathonn/ihavefood/api-gateway/genproto"
)

type gateway struct{}

func New() *gateway {
	return new(gateway)
}

func (g *gateway) SetupMux() *runtime.ServeMux {

	mars := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Indent:    "  ",
			Multiline: true, // Optional, implied by presence of "Indent".
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	cl, err := grpc.NewClient(fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT")), opt)
	if err != nil {
		log.Fatalf("failed to create new client: %v", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json+pretty", mars),
		runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(cl)),
	)

	// BUG: it does not panick when uri incorrect(cause: it only register not call.).
	//
	// TODO: impl healthcheck
	for env, f := range map[string]func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		"CUSTOMER_URI": pb.RegisterCustomerServiceHandlerFromEndpoint,
		"COUPON_URI":   pb.RegisterCouponServiceHandlerFromEndpoint,
		"ORDER_URI":    pb.RegisterOrderServiceHandlerFromEndpoint,
		"MERCHANT_URI": pb.RegisterMerchantServiceHandlerFromEndpoint,
		"DELIVERY_URI": pb.RegisterDeliveryServiceHandlerFromEndpoint,
		"AUTH_URI":     pb.RegisterAuthServiceHandlerFromEndpoint,
	} {
		uri := os.Getenv(env)

		if uri == "" {
			log.Fatalf("%s is not set", env)
		}
		if err := f(context.TODO(), gwmux, uri, []grpc.DialOption{opt}); err != nil {
			log.Fatalf("Failed to register %s: %v", env, err)
		}
	}

	return gwmux
}

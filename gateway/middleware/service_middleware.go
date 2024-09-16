// TODO might delete this file

package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
)

var (
	errRestaurantClosed = errors.New("restaurant closed")
)

type ServiceMiddleware interface {
	VerifyPlaceOrder(next http.Handler) http.Handler
}

// ServiceMiddlewareConfig holds the clients required for the serviceMiddleware.
// The struct is used to simplify the parameter list of the NewServiceMiddleware function,
// making it easier to manage and pass multiple gRPC clients.
type ServiceMiddlewareConfig struct {
	RestaurantClient pb.RestaurantServiceClient
	CouponClient     pb.CouponServiceClient
	OrderClient      pb.OrderServiceClient
	DeliveryClient   pb.DeliveryServiceClient
	UserClient       pb.UserServiceClient
}

// serviceMiddleware is responsible for handling validation, verification,
// and orchestration tasks across various microservices. It contains gRPC
// clients for interacting with various services. facilitating communication
// and data exchange between these services.
type serviceMiddleware struct {
	restaurantClient pb.RestaurantServiceClient
	couponClient     pb.CouponServiceClient
	orderClient      pb.OrderServiceClient
	deliveryClient   pb.DeliveryServiceClient
	UserClient       pb.UserServiceClient
}

func NewServiceMiddleware(cfg ServiceMiddlewareConfig) ServiceMiddleware {
	return &serviceMiddleware{
		restaurantClient: cfg.RestaurantClient,
		couponClient:     cfg.CouponClient,
		orderClient:      cfg.OrderClient,
		deliveryClient:   cfg.DeliveryClient,
		UserClient:       cfg.UserClient,
	}
}

// TODO improve doc
// verifyPlaceOrder verifies the place order request by ....
func (m *serviceMiddleware) VerifyPlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		// re-assign value to request body
		r.Body = io.NopCloser(bytes.NewReader(body))

		var order pb.PlaceOrderRequest
		if err := json.Unmarshal(body, &order); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal: %v", err), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

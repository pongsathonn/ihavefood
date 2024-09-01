package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pb "github.com/pongsathonn/ihavefood/gateway/genproto"
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

// verifyPlaceOrder verifies the place order request by checking menu availability.
func (m *serviceMiddleware) VerifyPlaceOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
			return
		}

		// re-assign body to body request
		r.Body = io.NopCloser(bytes.NewReader(body))

		var req pb.CheckAvailableMenuRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal: %v", err), http.StatusBadRequest)
			return
		}

		if req.RestaurantName == "" || len(req.Menus) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		availMenu, err := m.availableMenu(req.RestaurantName, req.Menus)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to verify: %v", err), http.StatusBadRequest)
			return
		}

		if !availMenu /*|| !availCoupon*/ {
			http.Error(w, "menu not available", http.StatusBadRequest)
			return
		}

		//TODO verify valid coupon

		next.ServeHTTP(w, r)
	})
}

// availableMenu checks if the menu items for a given restaurant are available.
func (m *serviceMiddleware) availableMenu(restauName string, menus []*pb.Menu) (bool, error) {

	req := &pb.CheckAvailableMenuRequest{
		RestaurantName: restauName,
		Menus:          menus,
	}
	check, err := m.restaurantClient.CheckAvailableMenu(context.TODO(), req)
	if err != nil {
		return false, err
	}

	// 0: available, 1: unavailable, 2: unknown
	checkNumber := check.Available.Number()
	if checkNumber != 0 {
		return false, fmt.Errorf("menu status not available with number: %d", checkNumber)
	}

	return true, nil
}

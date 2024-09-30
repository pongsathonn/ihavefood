package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"slices"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
)

// Example data ( Chaing Mai district )
var example = map[string]*pb.Point{
	"Mueang":    &pb.Point{Latitude: 18.7883, Longitude: 98.9853},
	"Hang Dong": &pb.Point{Latitude: 18.6870, Longitude: 98.8897},
	"San Sai":   &pb.Point{Latitude: 18.8578, Longitude: 99.0631},
	"Mae Rim":   &pb.Point{Latitude: 18.8998, Longitude: 98.9311},
	"Doi Saket": &pb.Point{Latitude: 18.8482, Longitude: 99.1403},
}

// Example data for riders
var riders = []*pb.Rider{
	{RiderId: "001", RiderName: "Messi", PhoneNumber: "+1234567890"},
	{RiderId: "002", RiderName: "Ronaldo", PhoneNumber: "+1987654321"},
	{RiderId: "003", RiderName: "Neymar", PhoneNumber: "+1654321897"},
	{RiderId: "004", RiderName: "Pogba", PhoneNumber: "+3334445555"},
	{RiderId: "005", RiderName: "Halaand", PhoneNumber: "+7778889999"},
}

// pickUpInfo have field same as *pb.PickupInfo
// but add error field use in this file only
type pickUpInfo struct {
	OrderId        string
	PickupCode     string
	PickupLocation *pb.Point
	Destination    *pb.Point
	Error          error
}

type DeliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	mu         sync.Mutex
	rabbitmq   RabbitMQ
	repository DeliveryRepository

	riderAcceptedCh chan *pb.AcceptOrderRequest
	orderPickupCh   chan *pickUpInfo
}

func NewDeliveryService(rb RabbitMQ, rp DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		rabbitmq:   rb,
		repository: rp,

		riderAcceptedCh: make(chan *pb.AcceptOrderRequest),
		orderPickupCh:   make(chan *pickUpInfo),
	}
}

// DeliveryAssignment handles incoming orders from the "order.placed.event"
// and saves the placed order to the database, finds the nearest riders, and
// notifies them. It waits for rider acceptance and responds with order
// pickup details if the order is accepted.
//
// When saving the order with SaveOrderDelivery, the order status remains
// "unaccepted" until a rider accepts the order.
func (x *DeliveryService) DeliveryAssignment() {

	for {
		placeOrder, err := x.receiveOrder("order.placed.event")
		if err != nil {
			slog.Error("receive new order", "err", err)
			return
		}
		go func(placeOrder *pb.PlaceOrder) {

			if placeOrder == nil {
				slog.Error("place order is empty")
				return
			}

			// waitRiderDuration is time for waiting rider accept order
			waitRiderDuration := time.Minute * 10
			ctx, cancel := context.WithTimeout(context.Background(), waitRiderDuration)
			defer cancel()

			// save new placeOrder to deliverydb (unaccept order)
			if err := x.repository.SaveOrderDelivery(ctx, placeOrder.OrderId); err != nil {
				slog.Error("save new order", "err", err)
				return
			}

			riders, err := calculateNearestRider(placeOrder.UserAddress, placeOrder.RestaurantAddress)
			if err != nil {
				slog.Error("calculate nearest riders", "err", err)
				return
			}

			orderPickup, err := x.generateOrderPickUp(placeOrder)
			if err != nil {
				slog.Error("generate order pickup", "err", err)
				return
			}

			if err := notifyToRider(ctx, riders, orderPickup); err != nil {
				slog.Error("notify to riders", "err", err)
				return
			}

			if err := x.waitForRiderAccept(ctx, riders, orderPickup); err != nil {
				slog.Error("waiting rider accept", "err", err)
				return
			}

		}(placeOrder)
	}

}

func (x *DeliveryService) TrackOrder(ctx context.Context, in *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {

	//TODO implement

	return nil, status.Error(codes.Unimplemented, "method TrackOrder not implemented")
}

// AcceptOrder handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (x *DeliveryService) AcceptOrder(ctx context.Context, in *pb.AcceptOrderRequest) (*pb.AcceptOrderResponse, error) {

	if in.OrderId == "" || in.RiderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order ID and rider ID must be provided")
	}

	order, err := x.repository.GetOrderDeliveryById(ctx, in.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrive order delivery %v", err)
	}

	if order.IsAccepted {
		return nil, status.Error(409, "order has already been accepted")
	}

	timeOut, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	select {
	// Notify that the rider has accepted the order
	case x.riderAcceptedCh <- &pb.AcceptOrderRequest{RiderId: in.RiderId, OrderId: in.OrderId}:
	case <-timeOut.Done():
		return nil, status.Error(codes.Internal, "failed to notify server that order accepted")
	}

	// Wait for the pickup information
	select {
	case order, ok := <-x.orderPickupCh:
		if !ok {
			return nil, status.Error(codes.Internal, "channel closed unexpectedly")
		}

		if order.Error != nil {
			slog.Error("retrive order pickup", "err", order.Error)
			return nil, status.Error(codes.Internal, "failed to retrieve order pickup information")
		}

		return &pb.AcceptOrderResponse{
			OrderId: in.OrderId,
			PickupInfo: &pb.PickupInfo{
				PickupCode:     order.PickupCode,
				PickupLocation: order.PickupLocation,
				Destination:    order.Destination,
			},
		}, nil

	case <-timeOut.Done():
		return nil, status.Error(codes.Internal, "context timeout while waiting for pickup information")

	}
}

func (x *DeliveryService) GetDeliveryFee(ctx context.Context, in *pb.GetDeliveryFeeRequest) (*pb.GetDeliveryFeeResponse, error) {
	if in.UserAddress == nil {
		return nil, status.Error(codes.InvalidArgument, "user address must be provided")
	}

	if in.RestaurantAddress == nil {
		return nil, status.Error(codes.InvalidArgument, "restaurant address must be provided")
	}

	deliveryFee, err := calculateDeliveryFee(in.RestaurantAddress, in.UserAddress)
	if err != nil {
		slog.Error("calculate delivery fee", "err", err)
		return nil, status.Error(codes.Internal, "failed to calculate delivery fee")
	}

	return &pb.GetDeliveryFeeResponse{DeliveryFee: deliveryFee}, nil
}

func (x *DeliveryService) ConfirmCashPayment(ctx context.Context, in *pb.ConfirmCashPaymentRequest) (*pb.ConfirmCashPaymentResponse, error) {
	if in.OrderId == "" || in.RiderId == "" {
		return nil, status.Error(codes.Internal, "order id or rider id must be provided")
	}

	res, err := x.repository.GetOrderDeliveryById(ctx, in.OrderId)
	if err != nil {
		slog.Error("retrive order delivery", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrive order delivery")
	}

	if res.RiderId != in.RiderId {
		return nil, status.Error(codes.InvalidArgument, "rider ID mismatch with the accepted rider")
	}

	const (
		exchange   = "delivery_exchange"
		routingKey = "order.paid.event"
	)

	body, err := json.Marshal(map[string]string{
		"orderId": res.OrderId,
		"riderId": res.RiderId,
	})
	if err != nil {
		slog.Error("marshal failed", "err", err)
	}

	err = x.rabbitmq.Publish(ctx,
		exchange,
		routingKey,
		[]byte(body),
	)
	if err != nil {
		slog.Error("publish failed", "routingkey", routingKey, "err", err)
	}

	return &pb.ConfirmCashPaymentResponse{Success: true}, nil
}

// receiveOrder subscribes to new order from OrderService
// and returns the received order.
func (x *DeliveryService) receiveOrder(routingKey string) (*pb.PlaceOrder, error) {

	deliveries, err := x.rabbitmq.Subscribe(
		nil,
		"order_exchange", // exchange
		"",               // queue
		routingKey,       // routing key
	)
	if err != nil {
		return nil, err
	}

	for delivery := range deliveries {
		var placeOrder pb.PlaceOrder
		if err := json.Unmarshal(delivery.Body, &placeOrder); err != nil {
			return nil, err
		}
		return &placeOrder, nil
	}
	return nil, nil
}

// generateOrderPickUp is a function that generate pickupcode and locations
// to riderthat not accept order yet.
func (x *DeliveryService) generateOrderPickUp(placeOrder *pb.PlaceOrder) (*pickUpInfo, error) {

	code := randomThreeDigits()

	startPoint, err := x.addressToPoint(placeOrder.RestaurantAddress)
	if err != nil {
		return nil, err
	}

	destinationPoint, err := x.addressToPoint(placeOrder.UserAddress)
	if err != nil {
		return nil, err
	}

	return &pickUpInfo{
		OrderId:        placeOrder.OrderId,
		PickupCode:     code,
		PickupLocation: startPoint,
		Destination:    destinationPoint,
	}, nil

}

// addressToPoint convert Address to Locations point. this function not implement
// actual Geocoding ( Google APIs ) yet . just response with example data
//
// TODO implememnt Geocoding ( Google APIs )
func (x *DeliveryService) addressToPoint(addr *pb.Address) (*pb.Point, error) {

	point, ok := example[addr.District]
	if !ok {
		return nil, fmt.Errorf("district %s invalid", addr.District)
	}

	return point, nil
}

// waitRiderAccept waits for a rider to accept the order.
// The function listens for a rider's acceptance from the riderAcceptedCh channel.
// Upon receiving an acceptance:
//   - It validates if the rider was notified.
//   - Cancels the context to stop further notifications to other riders.
//   - Sends the order pickup information to the orderPickupCh channel.
//
// If no rider accepts the order within 15 minutes, the function cancels the context and stops waiting.
func (x *DeliveryService) waitForRiderAccept(ctx context.Context, riders []*pb.Rider, orderPickup *pickUpInfo) error {

	var riderIds []string
	for _, rider := range riders {
		riderIds = append(riderIds, rider.RiderId)
	}

	for {
		select {
		case accepted := <-x.riderAcceptedCh:
			if accepted.OrderId == "" || accepted.RiderId == "" {
				x.orderPickupCh <- &pickUpInfo{Error: errors.New("rider id or order id is empty")}
				continue
			}

			if !slices.Contains(riderIds, accepted.RiderId) {
				x.orderPickupCh <- &pickUpInfo{Error: fmt.Errorf("rider id %s not notified", accepted.RiderId)}
				continue
			}

			x.mu.Lock()
			// Save order delivery after rider accepted
			err := x.repository.UpdateOrderDelivery(ctx, &MOrderDelivery{
				OrderId:    accepted.OrderId,
				RiderId:    accepted.RiderId,
				IsAccepted: true,
				PickupCode: orderPickup.PickupCode,
				PickupLocation: &MPoint{
					Latitude:  orderPickup.PickupLocation.Latitude,
					Longitude: orderPickup.PickupLocation.Longitude,
				},
				Destination: &MPoint{
					Latitude:  orderPickup.PickupLocation.Latitude,
					Longitude: orderPickup.PickupLocation.Longitude,
				},
			})
			if err != nil {
				return err
			}
			x.mu.Unlock()

			slog.Info("Rider has accepted the order",
				"rider_id", accepted.RiderId,
				"order_id", accepted.OrderId,
			)

			select {
			// response order pickup information to function AcceptOrder
			case x.orderPickupCh <- orderPickup:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

// calculateNearestRider calculates and returns a list of riders who are
// geographically closest to the user's location, within a certain radius.
//
// This function uses the user's address to determine the proximity of available
// riders based on their current location. The algorithm should take into account
// the distance between the user's address and the riders' locations and return
// a list of the nearest riders.
func calculateNearestRider(userAddr *pb.Address, restaurantAddr *pb.Address) ([]*pb.Rider, error) {

	// TODO:
	//   - Implement the logic to calculate the radius between the user's address and
	//     the riders' locations.
	//   - Use an actual distance calculation algorithm (e.g., Haversine formula or
	//     another geo-location method) to filter riders within the radius.
	//   - Return a list of riders that are closest to the user's location.

	// riders is example data for nearest riders
	return riders, nil
}

// notifyToRider will notify to all nearest riders
//
// TODO implement push notification and stop notify if context done
func notifyToRider(ctx context.Context, riders []*pb.Rider, orderPickup *pickUpInfo) error {

	var riderIds []string
	for _, r := range riders {
		riderIds = append(riderIds, r.RiderId)
	}

	notifyInfo := struct {
		riderIds   []string
		orderId    string
		pickupCode string
	}{
		riderIds:   riderIds,
		orderId:    orderPickup.OrderId,
		pickupCode: orderPickup.PickupCode,
	}

	// Example log notify
	log.Printf("[NOTIFY INFO]: %+v\n", notifyInfo)

	return nil
}

// randomThreeDigits generate 3 digits pickup code between 100 - 999 .
func randomThreeDigits() string {

	// Half-open interval
	// - [a,b) is include a, exclude b
	// - (a,b] is include b, exclude a
	//
	// rand.Intn(900) returns a number between 0 and 899.
	// Adding 100 shifts the range to [100, 999].
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(900) + 100
	return strconv.Itoa(n)
}

// NOTE : Address point is example data use district only
// others address fields will be ignored
//
// calculateDeliveryFee calculate distance from user's address to restaurant's address
func calculateDeliveryFee(userAddr *pb.Address, restaurantAddr *pb.Address) (int32, error) {

	point1, ok1 := example[userAddr.District]
	point2, ok2 := example[restaurantAddr.District]

	if !ok1 || !ok2 {
		validDistricts := []string{
			"Mueang",
			"Hang Dong",
			"San Sai",
			"Mae Rim",
			"Doi Saket",
		}
		return 0, fmt.Errorf("invalid distrct. valid districts are: %v ", validDistricts)
	}

	// distance in kilometers
	distance := haversineDistance(point1, point2)
	if distance < 0 || distance > 25 {
		return 0, errors.New("distance must be between 0 and 25 km")
	}

	var deliveryFee int32

	switch {
	case distance <= 5:
		deliveryFee = 0
	case distance <= 10:
		deliveryFee = 50
	default:
		deliveryFee = 100
	}

	return deliveryFee, nil
}

// haversineDistance calculates the distance between two geographic points in kilometers.
func haversineDistance(p1, p2 *pb.Point) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers.

	// Convert latitude and longitude from degrees to radians.
	lat1 := p1.Latitude * math.Pi / 180
	lon1 := p1.Longitude * math.Pi / 180
	lat2 := p2.Latitude * math.Pi / 180
	lon2 := p2.Longitude * math.Pi / 180

	// Calculate the distance using the Haversine formula.
	latSin := math.Sin(lat1) * math.Sin(lat2)
	latCos := math.Cos(lat1) * math.Cos(lat2) * math.Cos(lon2-lon1)
	distance := math.Acos(latSin+latCos) * earthRadius

	return distance
}

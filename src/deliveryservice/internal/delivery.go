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

type DeliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	// acceptc is a map of channels used to notify the
	// waitForRiderAccept function when a rider has accepted
	// an order. The key is the orderID, which uniquely
	// identifies each channel, and the value is the
	// channel for the riderID, indicating which rider has
	// accepted the order.
	acceptc map[string]chan *pb.AcceptOrderRequest

	// accepte is a map of channels used to communicate
	// any errors encountered during the execution of
	// waitForRiderAccept. The key is the orderID, and
	// the value is a channel for error responses.
	accepte map[string]chan error

	rabbitmq RabbitMQ

	repository DeliveryRepository

	mu sync.Mutex
}

func NewDeliveryService(rb RabbitMQ, rp DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		rabbitmq:   rb,
		repository: rp,

		acceptc: make(map[string]chan *pb.AcceptOrderRequest),
		accepte: make(map[string]chan error),
	}
}

func (x *DeliveryService) RunMessageProcessing() {

	// handle incoming order start delivery assignment
	go x.fetch("order.placed.event", x.deliveryAssignment())

	// notify to riders after restaurant cooking
	go x.fetch("order.cooking.event", nil)

	select {} // TODO use waitgroup instead
}

func (x *DeliveryService) fetch(routingKey string, messages chan<- []byte) {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"order_exchange", // exchange
		"",               // queue
		routingKey,       // routing key
	)
	if err != nil {
		slog.Error("subscribe failed", "err", err)
	}

	for delivery := range deliveries {
		messages <- delivery.Body
	}
}

// DeliveryAssignment handles incoming orders  and saves the placed order
// to the database, finds the nearest riders, and notifies them. It waits
// for rider acceptance and responds with order pickup details if the order
// is accepted.
//
// When saving the order with SaveOrderDelivery, the order status remains
// "unaccepted" until a rider accepts the order.
func (x *DeliveryService) deliveryAssignment() chan<- []byte {

	messages := make(chan []byte)

	for msg := range messages {
		var placeOrder pb.PlaceOrder
		if err := json.Unmarshal(msg, &placeOrder); err != nil {
			slog.Error("unmarshal failed", "err", err)
			continue
		}

		go func(p *pb.PlaceOrder) {

			if p == nil {
				slog.Error("place order is empty")
				return
			}

			riders, pickup, err := x.prepareOrderDelivery(p)
			if err != nil {
				slog.Error("prepare order delivery", "err", err)
				return
			}

			if err := notifyToRider(riders, pickup); err != nil {
				slog.Error("notify to riders", "err", err)
				return
			}

			// TODO publish "rider.notified.event"

			if err := x.waitForRiderAccept(
				p.No,
				riders,
				pickup,
			); err != nil {
				slog.Error("waiting rider accept", "err", err)
				return
			}

			// TODO publish "rider.accepted.event"

		}(&placeOrder)
	}

	return messages

}

// prepareOrderDelivery will  calculate the nearest riders and generate order pickup
// infomations and save new placeorder to database,
func (x *DeliveryService) prepareOrderDelivery(order *pb.PlaceOrder) ([]*pb.Rider, *pb.PickupInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	riders, err := calculateNearestRider(order.UserAddress, order.RestaurantAddress)
	if err != nil {
		return nil, nil, err
	}

	pickup, err := x.generateOrderPickUp(order)
	if err != nil {
		return nil, nil, err
	}

	err = x.repository.SaveDelivery(ctx, &DeliveryEntity{
		OrderNO:    order.No,
		PickupCode: pickup.PickupCode,
		PickupLocation: &Point{
			Latitude:  pickup.PickupLocation.Latitude,
			Longitude: pickup.PickupLocation.Longitude,
		},
		Destination: &Point{
			Latitude:  pickup.Destination.Latitude,
			Longitude: pickup.Destination.Longitude,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return riders, pickup, nil
}

func (x *DeliveryService) TrackOrder(ctx context.Context, in *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {

	//TODO implement

	return nil, status.Error(codes.Unimplemented, "method TrackOrder not implemented")
}

// AcceptOrder handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (x *DeliveryService) AcceptOrder(ctx context.Context, in *pb.AcceptOrderRequest) (*pb.AcceptOrderResponse, error) {

	if in.OrderNo == "" || in.RiderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order number and rider id must be provided")
	}

	order, err := x.repository.GetDelivery(ctx, in.OrderNo)
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
	case x.acceptc[in.OrderNo] <- &pb.AcceptOrderRequest{
		RiderId: in.RiderId,
		OrderNo: in.OrderNo,
	}:
	case <-timeOut.Done():
		return nil, status.Error(codes.Internal, "failed to notify server that order accepted")
	}

	// TODO wait for response error accepte

	return &pb.AcceptOrderResponse{
		PickupInfo: &pb.PickupInfo{
			OrderNo:    order.OrderNO,
			PickupCode: order.PickupCode,
			PickupLocation: &pb.Point{
				Latitude:  order.PickupLocation.Latitude,
				Longitude: order.PickupLocation.Longitude,
			},
			Destination: &pb.Point{
				Latitude:  order.Destination.Latitude,
				Longitude: order.Destination.Longitude,
			},
		},
	}, nil

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
	if in.OrderNo == "" || in.RiderId == "" {
		return nil, status.Error(codes.Internal, "order number or rider id must be provided")
	}

	res, err := x.repository.GetDelivery(ctx, in.OrderNo)
	if err != nil {
		slog.Error("retrive order delivery", "err", err)
		return nil, status.Error(codes.Internal, "failed to retrive order delivery")
	}

	if res.RiderID != in.RiderId {
		return nil, status.Error(codes.InvalidArgument, "rider ID mismatch with the accepted rider")
	}

	const (
		exchange   = "delivery_exchange"
		routingKey = "order.paid.event"
	)

	body, err := json.Marshal(map[string]string{
		"orderNO": res.OrderNO,
		"riderID": res.RiderID,
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

// generateOrderPickUp is a function that generate pickup code and locations
// to riderthat not accept order yet.
func (x *DeliveryService) generateOrderPickUp(order *pb.PlaceOrder) (*pb.PickupInfo, error) {

	code := randomThreeDigits()

	startPoint, err := x.addressToPoint(order.RestaurantAddress)
	if err != nil {
		return nil, err
	}

	destinationPoint, err := x.addressToPoint(order.UserAddress)
	if err != nil {
		return nil, err
	}

	return &pb.PickupInfo{
		OrderNo:        order.No,
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
func (x *DeliveryService) waitForRiderAccept(
	orderNO string,
	riders []*pb.Rider,
	pickup *pb.PickupInfo,
) error {

	var riderIDs []string
	for _, rider := range riders {
		riderIDs = append(riderIDs, rider.Id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	x.acceptc[orderNO] = make(chan *pb.AcceptOrderRequest)

	for {
		select {
		case v := <-x.acceptc[orderNO]:
			if v.OrderNo == "" || v.RiderId == "" {
				x.accepte[orderNO] <- errors.New("riderID or orderNO is empty")
				continue
			}

			if !slices.Contains(riderIDs, v.RiderId) {
				x.accepte[orderNO] <- fmt.Errorf("riderID %s was not notified", v.RiderId)
				continue
			}

			// Save order delivery after rider accepted
			err := x.repository.UpdateDelivery(ctx, &DeliveryEntity{
				OrderNO:    v.OrderNo,
				RiderID:    v.RiderId,
				IsAccepted: true,
			})
			if err != nil {
				slog.Error("update delivery", "err", err)
				continue
			}

			slog.Info("Rider has accepted the order",
				"orderNO", v.OrderNo,
				"riderID", v.RiderId,
			)

			select {
			case x.accepte[orderNO] <- nil:
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
// TODO implement push notification
func notifyToRider(riders []*pb.Rider, pickup *pb.PickupInfo) error {

	var riderIDs []string
	for _, rider := range riders {
		riderIDs = append(riderIDs, rider.Id)
	}

	notifyInfo := struct {
		orderNo    string
		riderIDs   []string
		pickupCode string
	}{
		orderNo:    pickup.OrderNo,
		riderIDs:   riderIDs,
		pickupCode: pickup.PickupCode,
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

// NOTE: Address point is example data use district field only
// others address fields will be ignored
//
// # TODO use actual data
//
// calculateDeliveryFee calculate distance from user's address to restaurant's address
func calculateDeliveryFee(userAddr *pb.Address, restaurantAddr *pb.Address) (int32, error) {

	userPoint, ok1 := example[userAddr.District]
	restaurantPoint, ok2 := example[restaurantAddr.District]

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
	distance := haversineDistance(userPoint, restaurantPoint)
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

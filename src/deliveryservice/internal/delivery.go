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
	{Id: "001", Name: "Messi", PhoneNumber: "0846851976"},
	{Id: "002", Name: "Ronaldo", PhoneNumber: "0987858487"},
	{Id: "003", Name: "Neymar", PhoneNumber: "0684321352"},
	{Id: "004", Name: "Pogba", PhoneNumber: "0868549858"},
	{Id: "005", Name: "Halaand", PhoneNumber: "0932515487"},
}

// TODO improve channel doc
// cookingC handle "order.cooking.event" to trigger fn notifyToRider
// to notify nearest riders
//
// acceptC for trig from fn AcceptOrder to fn waitForRiderAccept
// when rider has accepted the order
//
// orderPickupCh TODO
var (
	cookingC map[string]chan struct{}
	acceptC  map[string]chan *pb.AcceptOrderRequest
	pickupC  map[string]chan struct{}
)

// pickUpInfo have field same as *pb.PickupInfo
// but add error field use in this file only
type pickUpInfo struct {
	OrderNo        string
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
}

func NewDeliveryService(rb RabbitMQ, rp DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		rabbitmq:   rb,
		repository: rp,
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

		go func(order *pb.PlaceOrder) {

			if order == nil {
				slog.Error("place order is empty")
				return
			}

			riders, orderPickup, err := x.prepareOrderDelivery(order)
			if err != nil {
				slog.Error("prepare order delivery", "err", err)
				return
			}

			cookingC[order.No] = make(chan struct{})

			// notifyToRider is executed after "order.cooking.event"
			if err := notifyToRider(cookingC[order.No], riders, orderPickup); err != nil {
				slog.Error("notify to riders", "err", err)
				return
			}

			// TODO publish "rider.notified.event"

			acceptC[order.No] = make(chan struct{})

			if err := x.waitForRiderAccept(acceptC[order.No], riders, orderPickup); err != nil {
				slog.Error("waiting rider accept", "err", err)
				return
			}

			// TODO publish "rider.accepted.event"

		}(&placeOrder)
	}

	return messages

}

// prepareOrderDelivery will save new placeorder to database, calculate the nearest riders
// and generate order pickup infomations and return
func (x *DeliveryService) prepareOrderDelivery(order *pb.PlaceOrder) ([]*pb.Rider, *pickUpInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := x.repository.SaveOrderDelivery(ctx, order.No); err != nil {
		return nil, nil, err
	}

	riders, err := calculateNearestRider(order.UserAddress, order.RestaurantAddress)
	if err != nil {
		return nil, nil, err
	}

	orderPickup, err := x.generateOrderPickUp(order)
	if err != nil {
		return nil, nil, err
	}

	return riders, orderPickup, nil
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

// generateOrderPickUp is a function that generate pickupcode and locations
// to riderthat not accept order yet.
func (x *DeliveryService) generateOrderPickUp(order *pb.PlaceOrder) (*pickUpInfo, error) {

	code := randomThreeDigits()

	startPoint, err := x.addressToPoint(order.RestaurantAddress)
	if err != nil {
		return nil, err
	}

	destinationPoint, err := x.addressToPoint(order.UserAddress)
	if err != nil {
		return nil, err
	}

	return &pickUpInfo{
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
	accepted <-chan *pb.AcceptOrderRequest,
	riders []*pb.Rider,
	orderPickup *pickUpInfo,
) error {

	var riderIds []string
	for _, rider := range riders {
		riderIds = append(riderIds, rider.RiderId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	for {
		select {
		case v := <-accepted:
			if v.OrderNo == "" || v.RiderId == "" {
				x.orderPickupCh <- &pickUpInfo{
					Error: errors.New("rider id or order number is empty"),
				}
				continue
			}

			if !slices.Contains(riderIds, v.RiderId) {
				x.orderPickupCh <- &pickUpInfo{Error: fmt.Errorf("rider id %s not notified", v.RiderId)}
				continue
			}

			x.mu.Lock()
			// Save order delivery after rider accepted
			err := x.repository.UpdateOrderDelivery(ctx, &MOrderDelivery{
				OrderNo:    v.OrderNo,
				RiderId:    v.RiderId,
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
				"orderNo", v.OrderNo,
				"riderId", v.RiderId,
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

// notifyToRider will notify to all nearest riders return nil after notified
//
// TODO implement push notification
func notifyToRider(
	cooking <-chan struct{},
	riders []*pb.Rider,
	orderPickup *pickUpInfo,
) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var riderIds []string
	for _, r := range riders {
		riderIds = append(riderIds, r.RiderId)
	}

	select {
	case <-cooking:

		notifyInfo := struct {
			riderIds   []string
			orderNo    string
			pickupCode string
		}{
			riderIds:   riderIds,
			orderNo:    orderPickup.OrderNo,
			pickupCode: orderPickup.PickupCode,
		}

		// Example log notify
		log.Printf("[NOTIFY INFO]: %+v\n", notifyInfo)

	case <-ctx.Done():
		return ctx.Err()
	}

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

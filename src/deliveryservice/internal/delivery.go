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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DeliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	// acceptc is a map of channels used to notify the
	// waitForRiderAccept function when a rider has accepted
	// an order. The key is the orderID, which uniquely
	// identifies each channel, and the value is the
	// channel for the riderID, indicating which rider has
	// accepted the order.
	acceptc map[string]chan *pb.ConfirmRiderAcceptRequest

	// accepte is a map of channels used to communicate
	// any errors encountered during the execution of
	// waitForRiderAccept. The key is the orderID, and
	// the value is a channel for error responses.
	accepte map[string]chan error

	rabbitmq RabbitMQ
	storage  DeliveryStorage
	mu       sync.Mutex
}

func NewDeliveryService(rabbitmq RabbitMQ, storage DeliveryStorage) *DeliveryService {
	return &DeliveryService{
		rabbitmq: rabbitmq,
		storage:  storage,

		acceptc: make(map[string]chan *pb.ConfirmRiderAcceptRequest),
		accepte: make(map[string]chan error),
	}
}

// GetOrderTracking response current rider location and
// timestamp every 1 minute. if delivery status is "DELIVRED".
// It will use deliverTime for timestamp and stop rider location.
func (x *DeliveryService) GetOrderTracking(in *pb.GetOrderTrackingRequest, stream pb.DeliveryService_GetOrderTrackingServer) error {

	var timestamp time.Time

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			delivery, err := x.storage.Delivery(context.TODO(), in.OrderId)
			if err != nil {
				slog.Error("failed to retrives delivery", "err", err)
				return status.Errorf(codes.Internal, "failed to retrives delivery: %v", err)
			}

			timestamp = time.Now()

			if delivery.Status == DELIVERED {
				timestamp = delivery.Timestamp.DeliverTime
			}

			err = stream.Send(&pb.GetOrderTrackingResponse{
				OrderId: delivery.OrderID,
				RiderId: delivery.RiderID,
				RiderLocation: &pb.Point{
					Latitude:  delivery.RiderLocation.Latitude,
					Longitude: delivery.RiderLocation.Longitude,
				},

				UpdateTime: timestamppb.New(timestamp),
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to response stream: %v", err)
			}

			// stop response if order delivered
			if delivery.Status == DELIVERED {
				break
			}
		}
	}

	return nil

}

func (x *DeliveryService) GetDeliveryFee(ctx context.Context, in *pb.GetDeliveryFeeRequest) (*pb.GetDeliveryFeeResponse, error) {

	//TODO validate input

	//TODO get restaurant address , user address

	restaurantPoint := &pb.Point{Latitude: in.RestaurantLat, Longitude: in.RestaurantLong}
	userPoint := &pb.Point{Latitude: in.UserLat, Longitude: in.UserLong}

	deliveryFee, err := calculateDeliveryFee(restaurantPoint, userPoint)
	if err != nil {
		slog.Error("calculate delivery fee", "err", err)
		return nil, status.Error(codes.Internal, "failed to calculate delivery fee")
	}

	return &pb.GetDeliveryFeeResponse{DeliveryFee: deliveryFee}, nil
}

// ConfirmRiderAccept handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (x *DeliveryService) ConfirmRiderAccept(ctx context.Context, in *pb.ConfirmRiderAcceptRequest) (*pb.PickupInfo, error) {

	if in.OrderId == "" || in.RiderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order number and rider id must be provided")
	}

	delivery, err := x.storage.Delivery(ctx, in.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrive order delivery %v", err)
	}

	if delivery.Status == DELIVERED {
		return nil, status.Error(409, "order has already been accepted")
	}

	timeOut, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	select {
	// Notify that the rider has accepted the order
	case x.acceptc[in.OrderId] <- &pb.ConfirmRiderAcceptRequest{
		RiderId: in.RiderId,
		OrderId: in.OrderId,
	}:
	case <-timeOut.Done():
		return nil, status.Error(codes.Internal, "failed to notify server that order accepted")
	}

	return &pb.PickupInfo{
		PickupCode: delivery.PickupCode,
		PickupLocation: &pb.Point{
			Latitude:  delivery.PickupLocation.Latitude,
			Longitude: delivery.PickupLocation.Longitude,
		},
		Destination: &pb.Point{
			Latitude:  delivery.Destination.Latitude,
			Longitude: delivery.Destination.Longitude,
		},
	}, nil

}

func (x *DeliveryService) ConfirmOrderDeliver(ctx context.Context, in *pb.ConfirmOrderDeliverRequest) (*emptypb.Empty, error) {

	if in.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order number must be provided")
	}

	if _, err := x.storage.UpdateStatus(ctx, in.OrderId, DELIVERED); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update delivery status %v", err)
	}

	return &emptypb.Empty{}, nil
}

// -------------------------------------------------------------------------------------------------------

func (x *DeliveryService) StartConsume() {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"rider.assign.queue", // queue
		"order.placed.event", // routing key
	)

	if err != nil {
		slog.Error("subscribe failed", "err", err)
	}

	for msg := range deliveries {
		x.deliveryAssignment() <- msg
	}

	select {}
}

// DeliveryAssignment handles incoming orders  and saves the placed order
// to the database, finds the nearest riders, and notifies them. It waits
// for rider acceptance and responds with order pickup details if the order
// is accepted.
//
// When saving the order with SaveOrderDelivery, the order status remains
// "unaccepted" until a rider accepts the order.
//
// messages is channel for receive new orders.
func (x *DeliveryService) deliveryAssignment() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	for msg := range messages {

		var placeOrder pb.PlaceOrder

		if err := json.Unmarshal(msg.Body, &placeOrder); err != nil {
			slog.Error("unmarshal failed", "err", err)
			continue
		}

		go func(order *pb.PlaceOrder) {

			if order == nil {
				slog.Error("place order is empty")
				return
			}

			riders, pickup, err := x.prepareOrderDelivery(order)
			if err != nil {
				slog.Error("prepare order delivery", "err", err)
				return
			}

			if err := notifyToRider(riders, pickup); err != nil {
				slog.Error("notify to riders", "err", err)
				return
			}

			if err := x.waitForRiderAccept(order.OrderId, riders, pickup); err != nil {
				slog.Error("waiting rider accept", "err", err)
				return
			}

			err = x.rabbitmq.Publish(context.TODO(),
				"rider.assigned.event",
				amqp.Publishing{
					Type: "string",
					Body: []byte(order.OrderId),
				},
			)
			if err != nil {
				slog.Error("publish event", "err", err)
				return
			}

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

	now := time.Now()

	orderNO, err := x.storage.Create(ctx, &newDelivery{
		OrderID:    order.OrderId,
		PickupCode: pickup.PickupCode,
		PickupLocation: &dbPoint{
			Latitude:  pickup.PickupLocation.Latitude,
			Longitude: pickup.PickupLocation.Longitude,
		},
		Destination: &dbPoint{
			Latitude:  pickup.Destination.Latitude,
			Longitude: pickup.Destination.Longitude,
		},
		Status: UNACCEPT,
		Timestamp: &dbTimeStamp{
			CreateTime:  now,
			AcceptTime:  time.Time{},
			DeliverTime: time.Time{},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	slog.Info("create new order", "NO", orderNO)

	return riders, pickup, nil
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
		riderIDs   []string
		pickupCode string
	}{
		riderIDs:   riderIDs,
		pickupCode: pickup.PickupCode,
	}

	// Example log notify
	log.Printf("[NOTIFY INFO]: %+v\n", notifyInfo)

	return nil
}

// waitRiderAccept waits for a rider to accept the order.
// The function listens for a rider's acceptance from the riderAcceptedCh channel.
// Upon receiving an acceptance:
//   - It validates if the rider was notified.
//   - Cancels the context to stop further notifications to other riders.
//   - Sends the order pickup information to the orderPickupCh channel.
//
// If no rider accepts the order within 15 minutes, the function cancels the context and stops waiting.
func (x *DeliveryService) waitForRiderAccept(orderNO string, riders []*pb.Rider,
	pickup *pb.PickupInfo) error {

	var riderIDs []string
	for _, rider := range riders {
		riderIDs = append(riderIDs, rider.Id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	x.acceptc[orderNO] = make(chan *pb.ConfirmRiderAcceptRequest)

	for {
		select {
		case v := <-x.acceptc[orderNO]:
			if v.OrderId == "" || v.RiderId == "" {
				x.accepte[orderNO] <- errors.New("riderID or orderNO is empty")
				continue
			}

			if !slices.Contains(riderIDs, v.RiderId) {
				x.accepte[orderNO] <- fmt.Errorf("riderID %s was not notified", v.RiderId)
				continue
			}

			// Save order delivery after rider accepted
			if _, err := x.storage.UpdateRiderAccept(ctx, v.OrderId, v.RiderId); err != nil {
				slog.Error("update delivery", "err", err)
				continue
			}

			slog.Info("Rider has accepted the order",
				"orderID", v.OrderId,
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

	// [ Chaing Mai district ]
	var example = map[string]*pb.Point{
		"Mueang":    &pb.Point{Latitude: 18.7883, Longitude: 98.9853},
		"Hang Dong": &pb.Point{Latitude: 18.6870, Longitude: 98.8897},
		"San Sai":   &pb.Point{Latitude: 18.8578, Longitude: 99.0631},
		"Mae Rim":   &pb.Point{Latitude: 18.8998, Longitude: 98.9311},
		"Doi Saket": &pb.Point{Latitude: 18.8482, Longitude: 99.1403},
	}

	point, ok := example[addr.District]
	if !ok {
		return nil, fmt.Errorf("district %s invalid", addr.District)
	}

	return point, nil
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

// calculateDeliveryFee calculate distance from user's address to restaurant's address
//
// TODO
// Calculates the delivery fee based on the distance between the restaurant and the user,
// using the road network for routing instead of a straight-line (point A to B) distance.
func calculateDeliveryFee(userPoint *pb.Point, restaurantPoint *pb.Point) (int32, error) {

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

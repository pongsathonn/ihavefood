package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
)

// pickUpInfo have field same as *pb.PickupInfo
// but add error field use in this file only
type pickUpInfo struct {
	PickupCode     string
	PickupLocation *pb.Point
	Destination    *pb.Point
	Error          error
}

// delivery implements the DeliveryServiceServer interface from the protobuf definition.
// Embed the unimplemented server for forward compatibility
// RabbitMQ pub/sub interface for message handling
// DeliveryRepo interface for data access
// riderAcceptedCh is used to send notifications about riders who have accepted an order.
// orderPickupCh is used to receive pickup information order
// notifiedRidersCh  used to send Riders list that notified
type DeliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	mu               sync.Mutex
	rabbitmq         RabbitMQ
	repository       DeliveryRepository
	restaurantClient pb.RestaurantServiceClient

	riderAcceptedCh chan *pb.AcceptOrderRequest
	orderPickupCh   chan *pickUpInfo
}

// NewDeliveryServer creates and initializes a new delivery instance.
func NewDeliveryService(rb RabbitMQ, rp DeliveryRepository, rc pb.RestaurantServiceClient) *DeliveryService {
	return &DeliveryService{
		rabbitmq:         rb,
		repository:       rp,
		restaurantClient: rc,

		riderAcceptedCh: make(chan *pb.AcceptOrderRequest),
		orderPickupCh:   make(chan *pickUpInfo),
	}
}

// TrackOrder handles requests for tracking an order. This method is not yet implemented.
func (x *DeliveryService) TrackOrder(ctx context.Context, in *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {

	//TODO implement

	return nil, status.Errorf(codes.Unimplemented, "method TrackOrder not implemented")
}

// AcceptOrder handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (x *DeliveryService) AcceptOrder(ctx context.Context, in *pb.AcceptOrderRequest) (*pb.AcceptOrderResponse, error) {

	// Validate the input request
	if in.OrderId == "" || in.RiderId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "order ID and rider ID must be provided")
	}

	order, err := x.repository.GetOrderDeliveryById(ctx, in.OrderId)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.NotFound, "order not found: %v ", err)
	}

	if order.IsAccepted {
		return nil, status.Errorf(409, "order has already been accepted")
	}

	x.mu.Lock()
	defer x.mu.Unlock()

	// Notify that the rider has accepted the order
	x.riderAcceptedCh <- &pb.AcceptOrderRequest{RiderId: in.RiderId, OrderId: in.OrderId}

	// Wait for the pickup information or timeout after 10 seconds
	select {
	case order, ok := <-x.orderPickupCh:

		if !ok {
			return nil, status.Errorf(codes.Internal, "channel closed unexpectedly")
		}

		if order.Error != nil {
			log.Println(order.Error.Error)
			return nil, status.Errorf(codes.Internal, "failed to retrieve order pickup information: %s", order.Error.Error)
		}

		// Save order delivery after rider accept to database
		orderDelivery := &MOrderDelivery{
			OrderId:    in.OrderId,
			RiderId:    in.RiderId,
			IsAccepted: true,
			PickupCode: order.PickupCode,
			PickupLocation: &MPoint{
				Latitude:  order.PickupLocation.Latitude,
				Longitude: order.PickupLocation.Longitude,
			},
			Destination: &MPoint{
				Latitude:  order.PickupLocation.Latitude,
				Longitude: order.PickupLocation.Longitude,
			},
		}

		if err := x.repository.UpdateOrderDelivery(ctx, orderDelivery); err != nil {
			log.Println(err)
			return nil, status.Errorf(codes.Internal, "failed to save order delivery information")
		}

		pickupInfo := &pb.PickupInfo{
			PickupCode:     order.PickupCode,
			PickupLocation: order.PickupLocation,
			Destination:    order.Destination,
		}
		return &pb.AcceptOrderResponse{PickupInfo: pickupInfo}, nil

	case <-time.After(10 * time.Second):
		return nil, status.Errorf(codes.Internal, "timeout while waiting for pickup information")
	}
}

// OrderAssignment handles incoming orders, saves them to the database,
// calculates the nearest riders, and notifies them. It waits for rider
// acceptance and responds with order pickup details if accepted.
//
// when save order with SaveOrderDelivery the order status still be not accept
// after rider accept order it change order status at function AcceptOrder
func (x *DeliveryService) OrderAssignment() {

	for {
		placeOrder := x.receiveOrder()
		go func(placeOrder *pb.PlaceOrder) {

			// save new placeOrder to deliverydb ( not accepted yet )
			err := x.repository.SaveOrderDelivery(context.TODO(), placeOrder.OrderId)
			if err != nil {
				log.Printf("failed to save new order: %v", err)
				return
			}

			riders, err := x.calculateNearestRider(placeOrder.Address)
			if err != nil {
				log.Printf("failed to calculate nearest riders: %v", err)
				return
			}

			orderPickup, err := x.generateOrderPickUp(placeOrder)
			if err != nil {
				log.Println("failed to generate order pickup: %v", err)
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// notify to rider and wait for rider accept order
			// if rider is
			go x.waitRiderAcceptance(ctx, cancel, riders, orderPickup)
			x.notifyToRider(ctx, riders, orderPickup)

		}(placeOrder)
	}

}

// receiveOrder subscribes to new order from OrderService
// and returns the received order.
func (x *DeliveryService) receiveOrder() *pb.PlaceOrder {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"order_exchange",     // exchange
		"",                   // queue
		"order.placed.event", // routing key
	)
	if err != nil {
		log.Println("failed to subscrib to order queue: %v", err)
		return nil
	}

	for delivery := range deliveries {
		var placeOrder pb.PlaceOrder
		if err := json.Unmarshal(delivery.Body, &placeOrder); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return nil
		}
		return &placeOrder
	}
	return nil
}

// calculateNearestRider calculates and returns a list of riders nearest to the given address.
// This function needs implementation.
func (x *DeliveryService) calculateNearestRider(userAddr *pb.Address) ([]*pb.Rider, error) {

	// TODO
	// - Implement actual algorithm to calculate nearest riders based on user's address
	//   and convert user's address to latitude, longtitude

	// Example data for riders, used for testing or as mock data.
	riders := []*pb.Rider{
		{RiderId: "001", RiderName: "Messi", PhoneNumber: "+1234567890"},
		{RiderId: "002", RiderName: "Ronaldo", PhoneNumber: "+1987654321"},
		{RiderId: "003", RiderName: "Neymar", PhoneNumber: "+1654321897"},
		{RiderId: "004", RiderName: "Pogba", PhoneNumber: "+3334445555"},
		{RiderId: "005", RiderName: "Halaand", PhoneNumber: "+7778889999"},
	}
	return riders, nil
}

// generateOrderPickUp is a function that generate pickupcode and locations
// to riderthat not accept order yet.
func (x *DeliveryService) generateOrderPickUp(placeOrder *pb.PlaceOrder) (*pickUpInfo, error) {

	code := x.randomThreeDigits()

	req := &pb.GetRestaurantRequest{RestaurantName: placeOrder.RestaurantName}
	restaurant, err := x.restaurantClient.GetRestaurant(context.TODO(), req)
	if err != nil {
	}

	restaurantPoint, err := x.addressToPoint(restaurant.Restaurant.Address)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert restaurant address to point: %w", err)
	}

	destinationPoint, err := x.addressToPoint(placeOrder.Address)
	if err != nil {
		return nil, fmt.Errorf("couldn't convert destination address to point: %w", err)
	}

	return &pickUpInfo{
		PickupCode:     code,
		PickupLocation: restaurantPoint,
		Destination:    destinationPoint,
	}, nil

}

// addressToPoint use for convert address to locations point
func (x *DeliveryService) addressToPoint(addr *pb.Address) (*pb.Point, error) {

	// TODO implememnt Geocoding ( Google APIs )

	// Example data
	example := map[string]*pb.Point{
		"Bangkok":    &pb.Point{Latitude: 13.7563, Longitude: 100.5018},
		"Chiang Mai": &pb.Point{Latitude: 18.7883, Longitude: 98.9853},
		"Phuket":     &pb.Point{Latitude: 7.8804, Longitude: 98.3923},
		"Lampang":    &pb.Point{Latitude: 18.2888, Longitude: 99.4931},
		"Rayong":     &pb.Point{Latitude: 12.6828, Longitude: 101.2753},
	}

	point, ok := example[addr.Province]
	if !ok {
		return nil, fmt.Errorf("province %s not found", addr.Province)
	}

	return point, nil
}

// waitRiderAcceptance waits for a rider to accept the order.
// The function listens for a rider's acceptance from the riderAcceptedCh channel.
// Upon receiving an acceptance:
//   - It validates if the rider was notified.
//   - Cancels the context to stop further notifications to other riders.
//   - Sends the order pickup information to the orderPickupCh channel.
//
// If no rider accepts the order within 15 minutes, the function cancels the context and stops waiting.
func (x *DeliveryService) waitRiderAcceptance(ctx context.Context, cancel context.CancelFunc, riders []*pb.Rider, orderPickup *pickUpInfo) {

	var ridersId []string
	for _, rider := range riders {
		ridersId = append(ridersId, rider.RiderId)
	}

	select {
	case req := <-x.riderAcceptedCh:

		//  check rider is rider notified
		if !slices.Contains(ridersId, req.RiderId) {

			err := fmt.Errorf("rider %s not notified", req.RiderId)
			log.Println(err)

			x.orderPickupCh <- &pickUpInfo{Error: err}
			return
		}

		log.Printf("rider %s has accepted order with order pickup code %s", req.RiderId, orderPickup.PickupCode)

		// Call this cancel function to signal the context in the notifyToRider function.
		// This will stop any further notifications to other riders.
		cancel()

		// sending orderpickup information to function AcceptOrder
		x.orderPickupCh <- orderPickup

	case <-time.After(15 * time.Minute):
		cancel()
		return
	}
}

// randomThreeDigits generate 3 digits pickup code between 100 - 999 .
func (x *DeliveryService) randomThreeDigits() string {

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

// notifyToRider will notify to all nearest riders
func (x *DeliveryService) notifyToRider(ctx context.Context, riders []*pb.Rider, orderPickup *pickUpInfo) {

	// example notify
	log.Printf("started notify order %s", orderPickup.PickupCode)

	for _ = range riders {

		// TODO implement Push notification here

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			continue
		}

	}

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/pongsathonn/ihavefood/src/deliveryservice/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/deliveryservice/genproto"
)

// PickupInfo has same field with AcceptOrderHandlerResponse
// but only use in program
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
type deliveryService struct {
	pb.UnimplementedDeliveryServiceServer

	mu         sync.Mutex
	rabbitmq   RabbitmqClient
	repository repository.DeliveryRepo

	riderAcceptedCh chan *pb.AcceptOrderHandlerRequest
	orderPickupCh   chan *pickUpInfo
}

// NewDeliveryServer creates and initializes a new delivery instance.
func NewDeliveryService(rabbitmq RabbitmqClient, repository repository.DeliveryRepo) *deliveryService {
	return &deliveryService{
		rabbitmq:   rabbitmq,
		repository: repository,

		riderAcceptedCh: make(chan *pb.AcceptOrderHandlerRequest),
		orderPickupCh:   make(chan *pickUpInfo),
	}
}

// TrackOrder handles requests for tracking an order. This method is not yet implemented.
func (x *deliveryService) TrackOrder(ctx context.Context, in *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {

	//TODO implement

	return nil, status.Errorf(codes.Unimplemented, "method TrackOrder not implemented")
}

// AcceptOrderHandler handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (x *deliveryService) AcceptOrderHandler(ctx context.Context, in *pb.AcceptOrderHandlerRequest) (*pb.AcceptOrderHandlerResponse, error) {

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
	x.riderAcceptedCh <- &pb.AcceptOrderHandlerRequest{RiderId: in.RiderId, OrderId: in.OrderId}

	// Save the order delivery information to the database
	if err := x.repository.UpdateOrderDelivery(ctx, in.OrderId, in.RiderId, true); err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to save order delivery information")
	}

	// Wait for the pickup information or timeout after 30 seconds
	select {
	case order, ok := <-x.orderPickupCh:

		if !ok {
			return nil, status.Errorf(codes.Internal, "channel closed unexpectedly")
		}

		if order.Error != nil {
			log.Println(order.Error.Error)
			return nil, status.Errorf(codes.Internal, "failed to retrieve order pickup information: %s", order.Error.Error)
		}

		return &pb.AcceptOrderHandlerResponse{
			PickupCode:     order.PickupCode,
			PickupLocation: order.PickupLocation,
			Destination:    order.Destination,
		}, nil

	case <-time.After(30 * time.Second):
		return nil, status.Errorf(codes.Internal, "timeout while waiting for pickup information")
	}
}

// orderAssignment is responsible for receiving orders and assign them to riders.
func (x *deliveryService) orderAssignment() {

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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			//TODO generateOrderPickup receive input placeOrder
			orderPickup, err := x.generateOrderPickUp()
			if err != nil {
				log.Println("failed to generate order pickup: %v", err)
				return
			}

			// waiting for rider accept order
			go x.waitRiderAcceptance(ctx, cancel, riders, orderPickup)
			x.notifyToRider(ctx, riders, orderPickup)

		}(placeOrder)
	}

}

// receiveOrder subscribes to the RabbitMQ queue and returns the received order.
func (x *deliveryService) receiveOrder() *pb.PlaceOrder {

	deliveries, err := x.rabbitmq.Subscribe(
		"order",              // exchange
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
func (x *deliveryService) calculateNearestRider(addr *pb.Address) ([]*pb.Rider, error) {

	// TODO: Implement algorithm to calculate nearest riders based on address

	// Example data for riders, typically used for testing or as mock data.
	riders := []*pb.Rider{
		{RiderId: "rider001", RiderName: "Suriya Jaidi", PhoneNumber: "+1234567890"},
		{RiderId: "rider002", RiderName: "Warinee Sukchai", PhoneNumber: "+1987654321"},
		{RiderId: "rider003", RiderName: "Cheenchom Prabussa", PhoneNumber: "+1654321897"},
		{RiderId: "rider004", RiderName: "Janchonchan Chomphu", PhoneNumber: "+3334445555"},
		{RiderId: "rider005", RiderName: "Sudarat Prasang", PhoneNumber: "+7778889999"},
	}

	return riders, nil
}

// This function needs implementation.
// receive order and calculate location and pickup code
func (x *deliveryService) generateOrderPickUp() (*pickUpInfo, error) {

	//TODO implement generate pickup code

	return &pickUpInfo{
		PickupCode:     "229",
		PickupLocation: &pb.Point{Latitude: "-1283712", Longtitude: "123120312"},
		Destination:    &pb.Point{Latitude: "-13123123", Longtitude: "91203820"},
	}, nil

}

// waitRiderAcceptance waiting for Rider nofitied accep order
func (x *deliveryService) waitRiderAcceptance(ctx context.Context, cancel context.CancelFunc, riders []*pb.Rider, orderPickup *pickUpInfo) {

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
		cancel()

		x.orderPickupCh <- orderPickup

	case <-time.After(15 * time.Minute):
		cancel()
	}
}

// notifyToRider will notify to all nearest riders
func (x *deliveryService) notifyToRider(ctx context.Context, riders []*pb.Rider, orderPickup *pickUpInfo) {

	log.Printf("started notify order %s", orderPickup.PickupCode)

	for _ = range riders {

		// TODO implement notify logic

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			continue
		}

	}

}

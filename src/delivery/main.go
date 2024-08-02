package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	amqp "github.com/rabbitmq/amqp091-go"

	pb "github.com/pongsathonn/ihavefood/src/delivery/genproto"
	"github.com/pongsathonn/ihavefood/src/delivery/pubsub"
	"github.com/pongsathonn/ihavefood/src/delivery/repository"
)

// PickupInfo has same field with AcceptOrderHandlerResponse
// but only use in program
type pickUpInfo struct {
	PickupCode     string
	PickupLocation *pb.Point
	Destination    *pb.Point
	Error          error
}

// deliveryServer implements the DeliveryServiceServer interface from the protobuf definition.
// Embed the unimplemented server for forward compatibility
// RabbitMQ pub/sub interface for message handling
// DeliveryRepo interface for data access
// riderAcceptedCh is used to send notifications about riders who have accepted an order.
// orderPickupCh is used to receive pickup information order
// notifiedRidersCh  used to send Riders list that notified
type deliveryServer struct {
	pb.UnimplementedDeliveryServiceServer

	mu sync.Mutex
	ps pubsub.RabbitMQ
	rp repository.DeliveryRepo

	riderAcceptedCh chan *pb.AcceptOrderHandlerRequest
	orderPickupCh   chan *pickUpInfo
}

// newDeliveryServer creates and initializes a new deliveryServer instance.
func newDeliveryServer(ps pubsub.RabbitMQ, rp repository.DeliveryRepo) *deliveryServer {
	return &deliveryServer{
		ps: ps,
		rp: rp,

		riderAcceptedCh: make(chan *pb.AcceptOrderHandlerRequest),
		orderPickupCh:   make(chan *pickUpInfo),
	}
}

// TrackOrder handles requests for tracking an order. This method is not yet implemented.
func (s *deliveryServer) TrackOrder(ctx context.Context, in *pb.TrackOrderRequest) (*pb.TrackOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TrackOrder not implemented")
}

// AcceptOrderHandler handles requests indicating that a rider has accepted an order.
// It saves the delivery order to the database. Each order should only be accepted once.
func (s *deliveryServer) AcceptOrderHandler(ctx context.Context, in *pb.AcceptOrderHandlerRequest) (*pb.AcceptOrderHandlerResponse, error) {

	// Validate the input request
	if in.OrderId == "" || in.RiderId == "" {
		return nil, errors.New("order ID and rider ID must be provided")
	}

	order, err := s.rp.GetOrderDeliveryById(ctx, in.OrderId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if order.IsAccepted {
		return nil, errors.New("order has already been accepted")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Notify that the rider has accepted the order
	s.riderAcceptedCh <- &pb.AcceptOrderHandlerRequest{RiderId: in.RiderId, OrderId: in.OrderId}

	// Save the order delivery information to the database
	if err := s.rp.UpdateOrderDelivery(ctx, in.OrderId, in.RiderId, true); err != nil {
		return nil, errors.New("failed to save order delivery information")
	}

	// Wait for the pickup information or timeout after 30 seconds
	select {
	case order, ok := <-s.orderPickupCh:

		if !ok {
			return nil, errors.New("channel closed unexpectedly")
		}

		if order.Error != nil {
			return nil, errors.New("internal error ja")
		}

		return &pb.AcceptOrderHandlerResponse{
			PickupCode:     order.PickupCode,
			PickupLocation: order.PickupLocation,
			Destination:    order.Destination,
		}, nil

	case <-time.After(30 * time.Second):
		return nil, errors.New("timeout while waiting for pickup information")
	}
}

// orderAssignment is responsible for receiving orders and assigning them to riders.
func (s *deliveryServer) orderAssignment() {

	for {

		placeOrder := s.receiveOrder()

		go func(placeOrder *pb.PlaceOrder) {

			// save new placeOrder to deliverydb ( not accepted yet )
			s.rp.SaveOrderDelivery(context.TODO(), placeOrder.OrderId)

			riders, err := s.calculateNearestRider(placeOrder.Address)
			if err != nil {
				log.Println("Error calculating nearest riders:", err)
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			//TODO generateOrderPickup receive input placeOrder
			orderPickup, err := s.generateOrderPickUp()
			if err != nil {
				log.Println("Error calculating order pickup:", err)
				return
			}

			// waiting for rider accept order
			go s.waitRiderAcceptance(ctx, cancel, riders, orderPickup)
			s.notifyToRider(ctx, riders, orderPickup)

		}(placeOrder)
	}

}

// receiveOrder subscribes to the RabbitMQ queue and returns the received order.
func (s *deliveryServer) receiveOrder() *pb.PlaceOrder {
	deliveries, err := s.ps.Subscribe()
	if err != nil {
		log.Println("Error subscribing to order queue:", err)
		return nil
	}

	for delivery := range deliveries {
		var placeOrder pb.PlaceOrder
		if err := json.Unmarshal(delivery.Body, &placeOrder); err != nil {
			log.Printf("Failed to unmarshal message: %w", err)
			return nil
		}
		return &placeOrder
	}

	return nil
}

// calculateNearestRider calculates and returns a list of riders nearest to the given address.
// This function needs implementation.
func (s *deliveryServer) calculateNearestRider(addr *pb.Address) ([]*pb.Rider, error) {

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
func (s *deliveryServer) generateOrderPickUp() (*pickUpInfo, error) {

	//TODO implement generate pickup code

	return &pickUpInfo{
		PickupCode:     "229",
		PickupLocation: &pb.Point{Latitude: "-1283712", Longtitude: "123120312"},
		Destination:    &pb.Point{Latitude: "-13123123", Longtitude: "91203820"},
	}, nil

}

// waitRiderAcceptance waiting for Rider nofitied accep order
func (s *deliveryServer) waitRiderAcceptance(ctx context.Context, cancel context.CancelFunc, riders []*pb.Rider, orderPickup *pickUpInfo) {

	var ridersId []string
	for _, rider := range riders {
		ridersId = append(ridersId, rider.RiderId)
	}

	select {
	case req := <-s.riderAcceptedCh:

		// TODO  check rider is rider notified
		if !slices.Contains(ridersId, req.RiderId) {
			log.Printf("we didn't notify this rider %s ", req.RiderId)
			s.orderPickupCh <- &pickUpInfo{Error: errors.New("invalid rider id")}
			return
		}

		log.Printf("rider %s has accepted order", req.RiderId)
		cancel()

		//response pickup order to rider
		s.orderPickupCh <- orderPickup

	case <-time.After(15 * time.Minute):
		cancel()
	}
}

// notifyToRider notify to all rider bla bla TODO fix doc
func (s *deliveryServer) notifyToRider(ctx context.Context, riders []*pb.Rider, orderPickup *pickUpInfo) {

	log.Printf("started notify order %s", orderPickup.PickupCode)

	for _ = range riders {

		// Assume this message is send to Rider
		// TODO implement notify function to all riders with order code

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			continue
		}

	}

}

// initPubSub initializes the RabbitMQ connection and returns the pubsub instance
func initPubSub() (pubsub.RabbitMQ, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%s",
		getEnv("DELIVERY_AMQP_USER", "donkadmin"),
		getEnv("DELIVERY_AMQP_PASS", "donkpassword"),
		getEnv("DELIVERY_AMQP_HOST", "localhost"),
		getEnv("DELIVERY_AMQP_PORT", "5672"),
	)

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return pubsub.NewRabbitMQ(conn), nil
}

// initRepository initializes the MongoDB connection and returns the delivery repository instance
func initRepository(ctx context.Context) (repository.DeliveryRepo, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/delivery_database?authSource=admin",
		getEnv("DELIVERY_MONGO_USER", "donkadmin"),
		getEnv("DELIVERY_MONGO_PASS", "donkpassword"),
		getEnv("DELIVERY_MONGO_HOST", "localhost"),
		getEnv("DELIVERY_MONGO_PORT", "27017"),
	)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return repository.NewDeliveryRepo(client), nil
}

// startGRPCServer sets up and starts the gRPC server
func startGRPCServer(ds *deliveryServer) {

	// Set up the server port from environment variable
	uri := fmt.Sprintf(":%s", getEnv("DELIVERY_SERVER_PORT", "5555"))
	lis, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	// Create and start the gRPC server
	s := grpc.NewServer()
	pb.RegisterDeliveryServiceServer(s, ds)

	log.Printf("Delivery service is running on port %s\n", getEnv("DELIVERY_SERVER_PORT", "5555"))

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// getEnv fetches an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependencies
	ps, err := initPubSub()
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}

	rp, err := initRepository(ctx)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB:", err)
	}

	ds := newDeliveryServer(ps, rp)

	// Start the order assignment process in a separate goroutine
	go ds.orderAssignment()

	// Set up and start the gRPC server
	startGRPCServer(ds)
}

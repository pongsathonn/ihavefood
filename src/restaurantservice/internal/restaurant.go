package internal

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
)

var (
	errNoRestaurantNumber = status.Error(codes.InvalidArgument, "restaurant number must be provided")
	errNoRestaurantName   = status.Error(codes.InvalidArgument, "restaurant name must be provided")
	errNoMenus            = status.Error(codes.InvalidArgument, "menu must be at least one")
	errNoFoodName         = status.Error(codes.InvalidArgument, "food name must be provided")
	errUnknownTypeMenu    = status.Error(codes.InvalidArgument, "menu status cannot be UNKNOWN")
)

type RestaurantService struct {
	pb.UnimplementedRestaurantServiceServer

	repository RestaurantRepository
	rabbitmq   RabbitMQ
}

func NewRestaurantService(repository RestaurantRepository, rabbitmq RabbitMQ) *RestaurantService {
	return &RestaurantService{
		repository: repository,
		rabbitmq:   rabbitmq,
	}
}

func (x *RestaurantService) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantNo == "" {
		return nil, errNoRestaurantNumber
	}

	restaurant, err := x.repository.Restaurant(ctx, in.RestaurantNo)
	if err != nil {
		slog.Error("retrive restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurant")
	}

	return &pb.GetRestaurantResponse{Restaurant: restaurant}, nil
}

func (x *RestaurantService) ListRestaurant(ctx context.Context, empty *pb.Empty) (*pb.ListRestaurantResponse, error) {

	restaurants, err := x.repository.Restaurants(ctx)
	if err != nil {
		slog.Error("retrive restaurants", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurants")
	}

	resp := &pb.ListRestaurantResponse{Restaurants: restaurants}

	return resp, nil
}

func (x *RestaurantService) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

	if in.RestaurantName == "" {
		return nil, errNoRestaurantName
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	no, err := x.repository.SaveRestaurant(ctx, in.RestaurantName, in.Menus, in.Address)
	if err != nil {
		slog.Error("save restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save restaurant")
	}

	return &pb.RegisterRestaurantResponse{RestaurantNo: no}, nil
}

func (x *RestaurantService) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.AddMenuResponse, error) {

	if in.RestaurantNo == "" {
		return nil, errNoRestaurantNumber
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	if err := x.repository.UpdateMenu(ctx, in.RestaurantNo, in.Menus); err != nil {
		slog.Error("update menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update menu in database")
	}

	return &pb.AddMenuResponse{Success: true}, nil
}

func (x *RestaurantService) OrderReady(ctx context.Context, in *pb.OrderReadyRequest) (*pb.OrderReadyResponse, error) {

	if in.OrderNo == "" || in.RestaurantNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order number or restaurant number must be provided")
	}

	bs, err := json.Marshal(in.OrderNo)
	if err != nil {
		slog.Error("marshal failed", "err", err)
	}

	//TODO might save order to database or not

	err = x.rabbitmq.Publish(
		ctx,
		"order_x",
		"food.ready.event",
		bs,
	)
	if err != nil {
		slog.Error("publish failed", "err", err)
	}

	slog.Info("published event",
		"orderNo", in.OrderNo,
		"restaurantNo", in.RestaurantNo,
		"routingKey", "food.ready.event",
	)

	return &pb.OrderReadyResponse{Success: true}, nil
}

// -------------------------------------------------------

func (x *RestaurantService) RunMessageProcessing(ctx context.Context) {

	go x.subPlaceOrder("order.placed.event", x.handlePlaceOrder())

	// go x.subTODO("order.TODOs.event", x.handleTODO())

	<-ctx.Done()
}

// subPlaceOrder consume deliveries from an exchange that
// binding wiht routingkey and sends to messages chan
func (x *RestaurantService) subPlaceOrder(routingKey string, orderCh chan<- *pb.PlaceOrder) {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"order_x",  // exchange
		"",         // queue
		routingKey, // routing key
	)
	if err != nil {
		slog.Error("subscribe order", "err", err)
	}

	for delivery := range deliveries {
		var placeOrder pb.PlaceOrder
		if err := json.Unmarshal(delivery.Body, &placeOrder); err != nil {
			slog.Error("unmarshal order", "err", err)
			continue
		}
		orderCh <- &placeOrder
	}
}

// handlePlaceOrder will notify to restaurant
// and wait for restaurant accept the order
// thne publish "restaurant.accepted.event"
func (x *RestaurantService) handlePlaceOrder() chan<- *pb.PlaceOrder {

	orderCh := make(chan *pb.PlaceOrder)

	go func() {
		for order := range orderCh {

			// Assume this logs notify to restaurant
			log.Printf("HI RESTAURANT! you have new order %s\n", order.No)

			body, err := json.Marshal(order.No)
			if err != nil {
				slog.Error("marshal", "err", err)
				continue
			}

			// TODO wait for restaurant accept here
			// <- AcceptOrder()

			// assume restaurant accept order after 10s
			time.Sleep(10 * time.Second)

			err = x.rabbitmq.Publish(
				context.TODO(),
				"order_x",
				"restaurant.accepted.event",
				body,
			)
			if err != nil {
				slog.Error("publish event", "err", err)
				continue
			}

			slog.Info("published event",
				"routingKey", "restaurant.accepted.event",
				"orderNo", order.No,
			)
		}
	}()

	return orderCh
}

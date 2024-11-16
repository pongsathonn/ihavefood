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
	amqp "github.com/rabbitmq/amqp091-go"
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

	storage  RestaurantStorage
	rabbitmq RabbitMQ
}

func NewRestaurantService(storage RestaurantStorage, rabbitmq RabbitMQ) *RestaurantService {
	return &RestaurantService{
		storage:  storage,
		rabbitmq: rabbitmq,
	}
}

func (x *RestaurantService) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.GetRestaurantResponse, error) {
	if in.RestaurantNo == "" {
		return nil, errNoRestaurantNumber
	}

	res, err := x.storage.Restaurant(ctx, in.RestaurantNo)
	if err != nil {
		slog.Error("failed to retrive restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurant")
	}

	var menus []*pb.Menu
	for _, m := range res.Menus {
		menus = append(menus, &pb.Menu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	return &pb.GetRestaurantResponse{
		Restaurant: &pb.Restaurant{
			No:    res.No.Hex(),
			Name:  res.Name,
			Menus: menus,
			Address: &pb.Address{
				AddressName: res.Address.AddressName,
				SubDistrict: res.Address.SubDistrict,
				District:    res.Address.District,
				Province:    res.Address.Province,
				PostalCode:  res.Address.PostalCode,
			},
			Status: pb.Status(res.Status),
		}}, nil
}

func (x *RestaurantService) ListRestaurant(ctx context.Context, empty *pb.Empty) (*pb.ListRestaurantResponse, error) {

	res, err := x.storage.Restaurants(ctx)
	if err != nil {
		slog.Error("failed to retrive restaurants", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurants")
	}

	var restaurants []*pb.Restaurant
	for _, restaurant := range res {

		var menus []*pb.Menu
		for _, menu := range restaurant.Menus {
			menus = append(menus, &pb.Menu{
				FoodName: menu.FoodName,
				Price:    menu.Price,
			})
		}

		restaurants = append(restaurants, &pb.Restaurant{
			No:    restaurant.No.Hex(),
			Name:  restaurant.Name,
			Menus: menus,
			Address: &pb.Address{
				AddressName: restaurant.Address.AddressName,
				SubDistrict: restaurant.Address.SubDistrict,
				District:    restaurant.Address.District,
				Province:    restaurant.Address.Province,
				PostalCode:  restaurant.Address.PostalCode,
			},
			Status: pb.Status(restaurant.Status),
		})
	}

	return &pb.ListRestaurantResponse{Restaurants: restaurants}, nil
}

func (x *RestaurantService) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.RegisterRestaurantResponse, error) {

	if in.RestaurantName == "" {
		return nil, errNoRestaurantName
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	restaurantNO, err := x.storage.SaveRestaurant(ctx, in.RestaurantName, in.Menus, in.Address)
	if err != nil {
		slog.Error("failed to save restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save restaurant")
	}

	return &pb.RegisterRestaurantResponse{RestaurantNo: restaurantNO}, nil
}

func (x *RestaurantService) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.AddMenuResponse, error) {

	if in.RestaurantNo == "" {
		return nil, errNoRestaurantNumber
	}

	if len(in.Menus) == 0 {
		return nil, errNoMenus
	}

	if err := x.storage.UpdateMenu(ctx, in.RestaurantNo, in.Menus); err != nil {
		slog.Error("failed to update menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update menu in database")
	}

	return &pb.AddMenuResponse{Success: true}, nil
}

func (x *RestaurantService) OrderReady(ctx context.Context, in *pb.OrderReadyRequest) (*pb.OrderReadyResponse, error) {

	if in.OrderNo == "" || in.RestaurantNo == "" {
		return nil, status.Error(codes.InvalidArgument, "order number or restaurant number must be provided")
	}

	//TODO might save order to database or not
	//x.storage.SaveRestaurant()

	err := x.rabbitmq.Publish(
		ctx,
		"food.ready.event",
		amqp.Publishing{Body: []byte(in.OrderNo)},
	)
	if err != nil {
		slog.Error("failed to publish ", "err", err)
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

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"restaurant_assign_queue",
		"order.placed.event",
	)
	if err != nil {
		slog.Error("failed to subscribe", "err", err)
	}

	for msg := range deliveries {
		x.handlePlaceOrder() <- msg
	}

	<-ctx.Done()
}

// handlePlaceOrder will notify to restaurant
// and wait for restaurant accept the order
// thne publish "restaurant.accepted.event"
func (x *RestaurantService) handlePlaceOrder() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	go func() {
		for msg := range messages {

			var order pb.PlaceOrder
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				slog.Error("failed to unmarshal", "err", err)
				continue
			}

			// assume this logs is push notification to restaurant
			log.Printf("HI RESTAURANT! you have new order %s\n", order.No)

			// TODO wait for restaurant accept here
			// <- AcceptOrder()

			// assume restaurant accept order after 10s
			time.Sleep(10 * time.Second)

			rk := "restaurant.accepted.event"
			err := x.rabbitmq.Publish(
				context.TODO(),
				rk,
				amqp.Publishing{
					Body: []byte(order.No),
				},
			)
			if err != nil {
				slog.Error("failed to publish event", "err", err)
				continue
			}

			slog.Info("published event",
				"routingKey", rk,
				"orderNo", order.No,
			)
		}
	}()

	return messages
}

package internal

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/restaurantservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
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

func (x *RestaurantService) GetRestaurant(ctx context.Context, in *pb.GetRestaurantRequest) (*pb.Restaurant, error) {

	// TODO validate input

	restaurant, err := x.storage.Restaurant(ctx, in.RestaurantId)
	if err != nil {
		slog.Error("failed to retrive restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurant")
	}
	return dbToProto(restaurant), nil
}

func (x *RestaurantService) ListRestaurant(ctx context.Context, empty *emptypb.Empty) (*pb.ListRestaurantResponse, error) {

	// TODO validate input

	results, err := x.storage.Restaurants(ctx)
	if err != nil {
		slog.Error("failed to retrive restaurants", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve restaurants")
	}

	var restaurants []*pb.Restaurant
	for _, result := range results {
		restaurant := dbToProto(result)
		restaurants = append(restaurants, restaurant)
	}

	return &pb.ListRestaurantResponse{Restaurants: restaurants}, nil
}

func (x *RestaurantService) RegisterRestaurant(ctx context.Context, in *pb.RegisterRestaurantRequest) (*pb.Restaurant, error) {

	// TODO validate input

	var menus []*dbMenu
	for _, m := range in.Menus {
		menus = append(menus, &dbMenu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	restaurantNO, err := x.storage.SaveRestaurant(ctx, &newRestaurant{
		RestaurantName: in.RestaurantName,
		Menus:          menus,
		address: &dbAddress{
			AddressName: in.Address.AddressName,
			SubDistrict: in.Address.SubDistrict,
			District:    in.Address.District,
			Province:    in.Address.Province,
			PostalCode:  in.Address.PostalCode,
		},
	})
	if err != nil {
		slog.Error("failed to save restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save restaurant")
	}

	restaurant, err := x.storage.Restaurant(ctx, restaurantNO)
	if err != nil {
		slog.Error("failed to retrive restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive restaurant")
	}

	return dbToProto(restaurant), nil
}

func (x *RestaurantService) AddMenu(ctx context.Context, in *pb.AddMenuRequest) (*pb.Restaurant, error) {

	// TODO validate input

	var menus []*dbMenu
	for _, m := range in.Menus {
		menus = append(menus, &dbMenu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	restaurantNO, err := x.storage.UpdateMenu(ctx, in.RestaurantId, menus)
	if err != nil {
		slog.Error("failed to update menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update menu in database")
	}

	restaurant, err := x.storage.Restaurant(ctx, restaurantNO)
	if err != nil {
		slog.Error("failed to retrive restaurant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive restaurant")
	}

	return dbToProto(restaurant), nil
}

func (x *RestaurantService) OrderReady(ctx context.Context, in *pb.OrderReadyRequest) (*emptypb.Empty, error) {

	// TODO validate input

	msg := amqp.Publishing{Body: []byte(in.OrderId)}
	if err := x.rabbitmq.Publish(ctx, "food.ready.event", msg); err != nil {
		slog.Error("failed to publish ", "err", err)
	}

	slog.Info("published event",
		"orderId", in.OrderId,
		"restaurantNo", in.RestaurantId,
		"routingKey", "food.ready.event",
	)

	return &emptypb.Empty{}, nil
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
			log.Printf("HI RESTAURANT! you have new order %s\n", order.OrderId)

			// TODO wait for restaurant accept here
			// <- AcceptOrder()

			// assume restaurant accept order after 10s
			time.Sleep(10 * time.Second)

			rk := "restaurant.accepted.event"
			err := x.rabbitmq.Publish(
				context.TODO(),
				rk,
				amqp.Publishing{
					Body: []byte(order.OrderId),
				},
			)
			if err != nil {
				slog.Error("failed to publish event", "err", err)
				continue
			}

			slog.Info("published event",
				"routingKey", rk,
				"orderId", order.OrderId,
			)
		}
	}()

	return messages
}

func dbToProto(restaurant *dbRestaurant) *pb.Restaurant {

	var menus []*pb.Menu
	for _, m := range restaurant.Menus {
		menus = append(menus, &pb.Menu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	return &pb.Restaurant{
		RestaurantId:   restaurant.No.Hex(),
		RestaurantName: restaurant.Name,
		Menus:          menus,
		Address: &pb.Address{
			AddressName: restaurant.Address.AddressName,
			SubDistrict: restaurant.Address.SubDistrict,
			District:    restaurant.Address.District,
			Province:    restaurant.Address.Province,
			PostalCode:  restaurant.Address.PostalCode,
		},
		Status: pb.RestaurantStatus(restaurant.Status),
	}
}

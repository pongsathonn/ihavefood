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

	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MerchantService struct {
	pb.UnimplementedMerchantServiceServer

	storage  MerchantStorage
	rabbitmq RabbitMQ
}

func NewMerchantService(storage MerchantStorage, rabbitmq RabbitMQ) *MerchantService {
	return &MerchantService{
		storage:  storage,
		rabbitmq: rabbitmq,
	}
}

func (x *MerchantService) GetMerchant(ctx context.Context, in *pb.GetMerchantRequest) (*pb.Merchant, error) {

	// TODO validate input

	merchant, err := x.storage.Merchant(ctx, in.MerchantId)
	if err != nil {
		slog.Error("failed to retrive merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve merchant")
	}
	return dbToProto(merchant), nil
}

func (x *MerchantService) ListMerchant(ctx context.Context, empty *emptypb.Empty) (*pb.ListMerchantsResponse, error) {

	// TODO validate input

	results, err := x.storage.Merchants(ctx)
	if err != nil {
		slog.Error("failed to retrive merchants", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve merchants")
	}

	var merchants []*pb.Merchant
	for _, result := range results {
		merchant := dbToProto(result)
		merchants = append(merchants, merchant)
	}

	return &pb.ListMerchantsResponse{Merchants: merchants}, nil
}

func (x *MerchantService) RegisterMerchant(ctx context.Context, in *pb.RegisterMerchantRequest) (*pb.Merchant, error) {

	// TODO validate input

	var menu []*dbMenuItem
	for _, m := range in.Menu {
		menu = append(menu, &dbMenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	merchantNO, err := x.storage.SaveMerchant(ctx, &newMerchant{
		MerchantName: in.MerchantName,
		Menu:         menu,
		Address: &dbAddress{
			AddressName: in.Address.AddressName,
			SubDistrict: in.Address.SubDistrict,
			District:    in.Address.District,
			Province:    in.Address.Province,
			PostalCode:  in.Address.PostalCode,
		},
	})
	if err != nil {
		slog.Error("failed to save merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save merchant")
	}

	merchant, err := x.storage.Merchant(ctx, merchantNO)
	if err != nil {
		slog.Error("failed to retrive merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive merchant")
	}

	return dbToProto(merchant), nil
}

func (x *MerchantService) AddMenu(ctx context.Context, in *pb.CreateMenuItemRequest) (*pb.Merchant, error) {

	// TODO validate input

	var menus []*dbMenuItem
	for _, m := range in.Menu {
		menus = append(menus, &dbMenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	merchantNO, err := x.storage.UpdateMenu(ctx, in.MerchantId, menus)
	if err != nil {
		slog.Error("failed to update menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update menu in database")
	}

	merchant, err := x.storage.Merchant(ctx, merchantNO)
	if err != nil {
		slog.Error("failed to retrive merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive merchant")
	}

	return dbToProto(merchant), nil
}

// -------------------------------------------------------

func (x *MerchantService) RunMessageProcessing(ctx context.Context) {

	deliveries, err := x.rabbitmq.Subscribe(
		context.TODO(),
		"merchant_assign_queue",
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

// handlePlaceOrder will notify to merchant
// and wait for merchant accept the order
// thne publish "merchant.accepted.event"
func (x *MerchantService) handlePlaceOrder() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	go func() {
		for msg := range messages {

			var order pb.PlaceOrder
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				slog.Error("failed to unmarshal", "err", err)
				continue
			}

			// assume this logs is push notification to merchant
			log.Printf("HI RESTAURANT! you have new order %s\n", order.OrderId)

			// TODO wait for merchant accept here
			// <- AcceptOrder()

			// assume merchant accept order after 10s
			time.Sleep(10 * time.Second)

			rk := "merchant.accepted.event"
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

func dbToProto(merchant *dbMerchant) *pb.Merchant {

	var menu []*pb.MenuItem
	for _, m := range merchant.Menu {
		menu = append(menu, &pb.MenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	return &pb.Merchant{
		MerchantId:   merchant.ID.Hex(),
		MerchantName: merchant.Name,
		Menu:         menu,
		Address: &pb.Address{
			AddressName: merchant.Address.AddressName,
			SubDistrict: merchant.Address.SubDistrict,
			District:    merchant.Address.District,
			Province:    merchant.Address.Province,
			PostalCode:  merchant.Address.PostalCode,
		},
		Status: pb.StoreStatus(merchant.Status),
	}
}

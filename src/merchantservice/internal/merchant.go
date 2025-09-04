package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/merchantservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MerchantStorage interface {
	GetMerchant(ctx context.Context, merchantID string) (*dbMerchant, error)
	ListMerchants(ctx context.Context) ([]*dbMerchant, error)
	SaveMerchant(ctx context.Context, merchantID string, merchantName string) (*dbMerchant, error)
	CreateMenu(ctx context.Context, merchantID string, menu []*dbMenuItem) ([]*dbMenuItem, error)
	UpdateMenuItem(ctx context.Context, merchantID string, updateMenu *dbMenuItem) (*dbMenuItem, error)
}

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

func (x *MerchantService) ListMerchant(ctx context.Context, empty *emptypb.Empty) (*pb.ListMerchantsResponse, error) {

	results, err := x.storage.ListMerchants(ctx)
	if err != nil {
		slog.Error("storage list merchants", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var merchants []*pb.Merchant
	for _, result := range results {
		merchant := dbToProto(result)
		merchants = append(merchants, merchant)
	}

	return &pb.ListMerchantsResponse{Merchants: merchants}, nil
}

func (x *MerchantService) GetMerchant(ctx context.Context, in *pb.GetMerchantRequest) (*pb.Merchant, error) {

	uuid, err := uuid.Parse(in.MerchantId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return nil, status.Error(codes.InvalidArgument, "uuid invalid for merchant id")
	}

	merchant, err := x.storage.GetMerchant(ctx, uuid.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		slog.Error("storage get merchant", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return dbToProto(merchant), nil
}

func (x *MerchantService) CreateMerchant(ctx context.Context, in *pb.CreateMerchantRequest) (*pb.Merchant, error) {

	uuid, err := uuid.Parse(in.MerchantId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return nil, status.Error(codes.InvalidArgument, "uuid invalid for merchant id")
	}

	merchant, err := x.storage.SaveMerchant(ctx, uuid.String(), in.MerchantName)
	if err != nil {
		slog.Error("storage save merchant", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return dbToProto(merchant), nil
}

func (x *MerchantService) UpdateMerchant(ctx context.Context, in *pb.UpdateMerchantRequest) (*pb.Merchant, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateMerchant not implemented")
}

func (x *MerchantService) CreateMenu(ctx context.Context, in *pb.CreateMenuRequest) (*pb.CreateMenuResponse, error) {

	uuid, err := uuid.Parse(in.MerchantId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return nil, status.Error(codes.InvalidArgument, "uuid invalid for merchant id")
	}

	var newMenu []*dbMenuItem
	for _, m := range in.GetNewMenu() {
		newMenu = append(newMenu, &dbMenuItem{
			FoodName:    m.FoodName,
			Price:       m.Price,
			Description: m.Description,
		})
	}

	createdMenu, err := x.storage.CreateMenu(ctx, uuid.String(), newMenu)
	if err != nil {
		slog.Error("storage create menu", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var menu []*pb.MenuItem
	for _, v := range createdMenu {
		menu = append(menu, &pb.MenuItem{
			ItemId:      v.ItemID,
			FoodName:    v.FoodName,
			Price:       v.Price,
			Description: v.Description,
			IsAvailable: v.IsAvailable,
		})
	}

	return &pb.CreateMenuResponse{Menu: menu}, nil
}

func (x *MerchantService) UpdateMenuItem(ctx context.Context, in *pb.UpdateMenuItemRequest) (*pb.MenuItem, error) {

	uuid, err := uuid.Parse(in.MerchantId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return nil, status.Error(codes.InvalidArgument, "uuid invalid for merchant id")
	}

	updatedMenu, err := x.storage.UpdateMenuItem(ctx, uuid.String(), &dbMenuItem{
		FoodName:    in.FoodName,
		Price:       in.Price,
		Description: in.Description,
		IsAvailable: in.IsAvailable,
	})
	if err != nil {
		slog.Error("storage update menu item", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.MenuItem{
		ItemId:      updatedMenu.ItemID,
		FoodName:    updatedMenu.FoodName,
		Price:       updatedMenu.Price,
		Description: updatedMenu.Description,
		IsAvailable: updatedMenu.IsAvailable,
	}, nil
}

func (x *MerchantService) UpdateStoreStatus(context.Context, *pb.UpdateStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateStoreStatus not implemented")
}

func (x *MerchantService) GetStoreStatus(context.Context, *pb.GetStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetStoreStatus not implemented")
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
			ItemId:      m.ItemID,
			FoodName:    m.FoodName,
			Price:       m.Price,
			Description: m.Description,
			IsAvailable: m.IsAvailable,
		})
	}

	var address *pb.Address
	if merchant.Address != nil {
		address = &pb.Address{
			AddressName: merchant.Address.AddressName,
			SubDistrict: merchant.Address.SubDistrict,
			District:    merchant.Address.District,
			Province:    merchant.Address.Province,
			PostalCode:  merchant.Address.PostalCode,
		}
	}

	return &pb.Merchant{
		MerchantId:   merchant.ID,
		MerchantName: merchant.Name,
		Menu:         menu,
		Address:      address,
		Status:       pb.StoreStatus(merchant.Status),
	}
}

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
	GetMerchant(ctx context.Context, merchantID string) (*DbMerchant, error)
	ListMerchants(ctx context.Context) ([]*DbMerchant, error)
	CreateMerchant(ctx context.Context, newMerchant *NewMerchant) (string, error)
	CreateMenu(ctx context.Context, merchantID string, menu []*DbMenuItem) ([]*DbMenuItem, error)
	UpdateMenuItem(ctx context.Context, merchantID string, updateMenu *DbMenuItem) (*DbMenuItem, error)
}

type MerchantService struct {
	pb.UnimplementedMerchantServiceServer

	Storage  MerchantStorage
	rabbitmq RabbitMQ
}

func NewMerchantService(storage MerchantStorage, rabbitmq RabbitMQ) *MerchantService {
	return &MerchantService{
		Storage:  storage,
		rabbitmq: rabbitmq,
	}
}

func (x *MerchantService) ListMerchants(ctx context.Context, empty *emptypb.Empty) (*pb.ListMerchantsResponse, error) {

	results, err := x.Storage.ListMerchants(ctx)
	if err != nil {
		slog.Error("storage list merchants", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var merchants []*pb.Merchant
	for _, result := range results {
		merchant := DbToProto(result)
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

	merchant, err := x.Storage.GetMerchant(ctx, uuid.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		slog.Error("storage get merchant", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return DbToProto(merchant), nil
}

func (x *MerchantService) CreateMerchant(ctx context.Context, in *pb.CreateMerchantRequest) (*pb.Merchant, error) {

	var newMerchant *NewMerchant
	id, err := x.Storage.CreateMerchant(ctx, newMerchant.FromProto(in))
	if err != nil {
		slog.Error("failed to create merchant", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	merchant, err := x.Storage.GetMerchant(ctx, id)
	if err != nil {
		slog.Error("failed to get merchant after creation", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return merchant.IntoProto(), nil
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

	var newMenu []*DbMenuItem
	for _, m := range in.GetNewMenu() {
		newMenu = append(newMenu, &DbMenuItem{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}

	createdMenu, err := x.Storage.CreateMenu(ctx, uuid.String(), newMenu)
	if err != nil {
		slog.Error("storage create menu", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var menu []*pb.MenuItem
	for _, v := range createdMenu {
		menu = append(menu, &pb.MenuItem{
			ItemId:   v.ItemID,
			FoodName: v.FoodName,
			Price:    v.Price,
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

	updatedMenu, err := x.Storage.UpdateMenuItem(ctx, uuid.String(), &DbMenuItem{
		FoodName: in.FoodName,
		Price:    in.Price,
	})
	if err != nil {
		slog.Error("storage update menu item", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.MenuItem{
		ItemId:   updatedMenu.ItemID,
		FoodName: updatedMenu.FoodName,
		Price:    updatedMenu.Price,
	}, nil
}

func (x *MerchantService) UpdateStoreStatus(context.Context, *pb.UpdateStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateStoreStatus not implemented")
}

func (x *MerchantService) GetStoreStatus(context.Context, *pb.GetStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetStoreStatus not implemented")
}

// -------------------------------------------------------

// TODO : create register for event handler
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

func DbToProto(merchant *DbMerchant) *pb.Merchant {

	var menu []*pb.MenuItem
	for _, dbItem := range merchant.Menu {
		menu = append(menu, &pb.MenuItem{
			ItemId:   dbItem.ItemID,
			FoodName: dbItem.FoodName,
			Price:    dbItem.Price,
			Image: &pb.ImageInfo{
				Url:  dbItem.ImageInfo.Url,
				Type: dbItem.ImageInfo.Type,
			},
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

	var imageInfo *pb.ImageInfo
	if merchant.ImageInfo != nil {
		imageInfo = &pb.ImageInfo{
			Url:  merchant.ImageInfo.Url,
			Type: merchant.ImageInfo.Type,
		}
	}

	var status pb.StoreStatus
	if v, ok := pb.StoreStatus_value[merchant.Status]; ok {
		status = pb.StoreStatus(v)
	}

	return &pb.Merchant{
		MerchantId:   merchant.ID,
		MerchantName: merchant.Name,
		Menu:         menu,
		Address:      address,
		Image:        imageInfo,
		Status:       status,
	}
}

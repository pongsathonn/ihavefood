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

type MerchantStorage interface {
	GetMerchant(ctx context.Context, merchantID string) (*dbMerchant, error)
	ListMerchants(ctx context.Context) ([]*dbMerchant, error)
	SaveMerchant(ctx context.Context, merchantID string) (*dbMerchant, error)
	UpdateMenu(ctx context.Context, merchantID string, menu []*dbMenuItem) ([]*dbMenuItem, error)
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

func (x *MerchantService) GetMerchant(ctx context.Context, in *pb.GetMerchantRequest) (*pb.Merchant, error) {

	merchant, err := x.storage.GetMerchant(ctx, in.MerchantId)
	if err != nil {
		slog.Error("failed to retrive merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve merchant")
	}
	return dbToProto(merchant), nil
}

func (x *MerchantService) CreateMerchant(ctx context.Context, in *pb.CreateMerchantRequest) (*pb.Merchant, error) {

	merchant, err := x.storage.SaveMerchant(ctx, in.MerchantId)
	if err != nil {
		slog.Error("failed to save merchant", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save merchant")
	}

	return dbToProto(merchant), nil
}

func (x *MerchantService) UpdateMerchant(ctx context.Context, in *pb.UpdateMerchantRequest) (*pb.Merchant, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMerchant not implemented")
}

func (x *MerchantService) UpdateMenu(ctx context.Context, in *pb.UpdateMenuRequest) (*pb.UpdateMenuResponse, error) {

	var menus []*dbMenuItem
	for _, m := range in.GetMenu() {
		menus = append(menus, &dbMenuItem{
			FoodName:    m.FoodName,
			Price:       m.Price,
			Description: m.Description,
			IsAvailable: m.IsAvailable,
		})
	}

	saveMenu, err := x.storage.UpdateMenu(ctx, in.GetMerchantId(), menus)
	if err != nil {
		slog.Error("failed to save menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to save menu in database")
	}

	var menu []*pb.MenuItem
	for _, v := range saveMenu {
		menu = append(menu, &pb.MenuItem{
			ItemId:      v.ItemID.Hex(),
			FoodName:    v.FoodName,
			Price:       v.Price,
			Description: v.Description,
			IsAvailable: v.IsAvailable,
		})
	}

	return &pb.UpdateMenuResponse{Menu: menu}, nil
}

func (x *MerchantService) UpdateMenuItem(ctx context.Context, in *pb.UpdateMenuItemRequest) (*pb.MenuItem, error) {

	updatedMenu, err := x.storage.UpdateMenuItem(ctx, in.MerchantId, &dbMenuItem{
		FoodName:    in.Menu.FoodName,
		Price:       in.Menu.Price,
		Description: in.Menu.Description,
		IsAvailable: in.Menu.IsAvailable,
	})
	if err != nil {
		slog.Error("failed to update menu", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update menu in database")
	}

	return &pb.MenuItem{
		ItemId:      updatedMenu.ItemID.Hex(),
		FoodName:    updatedMenu.FoodName,
		Price:       updatedMenu.Price,
		Description: updatedMenu.Description,
		IsAvailable: updatedMenu.IsAvailable,
	}, nil
}

func (x *MerchantService) UpdateStoreStatus(context.Context, *pb.UpdateStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStoreStatus not implemented")
}

func (x *MerchantService) GetStoreStatus(context.Context, *pb.GetStoreStatusRequest) (*pb.StoreStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStoreStatus not implemented")
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
			ItemId:      m.ItemID.Hex(),
			FoodName:    m.FoodName,
			Price:       m.Price,
			Description: m.Description,
			IsAvailable: m.IsAvailable,
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

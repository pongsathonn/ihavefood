package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CouponClient interface {
	RedeemCoupon(ctx context.Context, in *pb.RedeemCouponRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type CustomerClient interface {
	GetCustomer(ctx context.Context, in *pb.GetCustomerRequest, opts ...grpc.CallOption) (*pb.Customer, error)
}

type DeliveryClient interface {
	GetDeliveryFee(ctx context.Context, in *pb.GetDeliveryFeeRequest, opts ...grpc.CallOption) (*pb.GetDeliveryFeeResponse, error)
}

type MerchantClient interface {
	GetMerchant(ctx context.Context, in *pb.GetMerchantRequest, opts ...grpc.CallOption) (*pb.Merchant, error)
}

type OrderService struct {
	storage  OrderStorage
	rabbitmq RabbitMQ

	coupon   CouponClient
	customer CustomerClient
	delivery DeliveryClient
	merchant MerchantClient

	pb.UnimplementedOrderServiceServer
}

func NewOrderService(
	storage OrderStorage,
	rabbitmq RabbitMQ,
	coupon CouponClient,
	customer CustomerClient,
	delivery DeliveryClient,
	merchant MerchantClient,
) *OrderService {
	return &OrderService{
		storage:  storage,
		rabbitmq: rabbitmq,
		coupon:   coupon,
		customer: customer,
		delivery: delivery,
		merchant: merchant,
	}
}

func (x *OrderService) ListOrderHistory(ctx context.Context,
	in *pb.ListOrderHistoryRequest) (*pb.ListOrderHistoryResponse, error) {

	if in.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "ID must be provided")
	}

	dbOrders, err := x.storage.ListPlaceOrders(ctx, in.CustomerId)
	if err != nil {
		slog.Error("storage list place orders", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var placeOrders []*pb.PlaceOrder
	for _, dbOrder := range dbOrders {
		placeOrder := toProtoPlaceOrder(dbOrder)
		placeOrders = append(placeOrders, placeOrder)
	}

	return &pb.ListOrderHistoryResponse{PlaceOrders: placeOrders}, nil
}

func (x *OrderService) CreatePlaceOrder(ctx context.Context, in *pb.CreatePlaceOrderRequest) (*pb.PlaceOrder, error) {

	if err := ValidateStruct(in); err != nil {
		var ve myValidatorErrs
		if errors.As(err, &ve) {
			return nil, status.Errorf(codes.InvalidArgument, "failed to create place order: %s", ve.Error())
		}
		slog.Error("validate struct", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	newOrder, err := x.prepareNewOrder(in)
	if err != nil {
		slog.Error("Failed to prepare new order", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	orderID, err := x.storage.Create(ctx, newOrder)
	if err != nil {
		slog.Error("storage create new order", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var success bool
	defer func() {
		if !success {
			if err := x.storage.DeletePlaceOrder(ctx, orderID); err != nil {
				slog.Error("failed to cleanup order", "orderId", orderID, "err", err)
			}
		}
	}()

	dbOrder, err := x.storage.GetPlaceOrder(ctx, orderID)
	if err != nil {
		slog.Error("storage get place order", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	order := toProtoPlaceOrder(dbOrder)

	body, err := proto.Marshal(order)
	if err != nil {
		slog.Error("protobuf marshal failed", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	err = x.rabbitmq.Publish(ctx, "order.placed.event", amqp.Publishing{
		Type: "ihavefood.PlaceOrder",
		Body: body,
	})
	if err != nil {
		slog.Error("rabbitmq publish", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	success = true
	slog.Info("published event", "orderId", order.OrderId)
	return order, nil
}

//---------------------------------------------------------------------------------------

func (x *OrderService) StartConsume() {

	registerEvent := []struct {
		queue   string
		key     string
		handler chan<- amqp.Delivery
	}{
		{"order_status_update_queue", "rider.finding.event", x.handleOrderStatus()},
		{"order_status_update_queue", "merchant.accepted.event", x.handleOrderStatus()},
		{"order_status_update_queue", "rider.assigned.event", x.handleOrderStatus()},
		{"order_status_update_queue", "rider.delivered.event", x.handleOrderStatus()},
		{"payment_status_update_queue", "order.paid.event", x.handlePaymentStatus()},
	}

	for _, r := range registerEvent {
		go func(queue, key string, handler chan<- amqp.Delivery) {

			deliveries, err := x.rabbitmq.Subscribe(context.TODO(), queue, key)
			if err != nil {
				slog.Error("subscribe order", "err", err)
			}

			for delivery := range deliveries {
				handler <- delivery
			}

		}(r.queue, r.key, r.handler)
	}

	select {}
}

func (x *OrderService) handleOrderStatus() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)
	var status pb.OrderStatus

	go func() {
		for msg := range messages {

			switch msg.RoutingKey {
			case "rider.finding.event":
				status = pb.OrderStatus_ORDER_STATUS_FINDING_RIDER
			case "merchant.accepted.event":
				status = pb.OrderStatus_ORDER_STATUS_PREPARING_ORDER
			case "food.ready.event":
				status = pb.OrderStatus_ORDER_STATUS_WAIT_FOR_PICKUP
			case "rider.assigned.event":
				status = pb.OrderStatus_ORDER_STATUS_ONGOING
			case "rider.delivered.event":
				status = pb.OrderStatus_ORDER_STATUS_DELIVERED
			default:
				slog.Error("unknown routing key %s", "key", msg.RoutingKey)
				continue
			}

			var orderID string
			if err := json.Unmarshal(msg.Body, &orderID); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdateOrderStatus(context.TODO(),
				orderID, dbOrderStatus(status)); err != nil {
				slog.Error("updated status", "err", err, "orderId", orderID)
				continue
			}
		}
	}()
	return messages
}

// handlePaymentStatus updates payment status of an order after
// it has been successfully processed.
//   - Cash method update when rider received cash from customer after delivery.
//   - PromptPay and Credit card upon succussful transaction.
func (x *OrderService) handlePaymentStatus() chan<- amqp.Delivery {

	messages := make(chan amqp.Delivery)

	go func() {
		for msg := range messages {

			if msg.RoutingKey != "customer.paid.event" {
				slog.Error("unknown routing key ", "key", msg.RoutingKey)
				continue
			}

			var orderID string
			if err := json.Unmarshal(msg.Body, &orderID); err != nil {
				slog.Error("unmarshal failed", "err", err)
				continue
			}

			if _, err := x.storage.UpdatePaymentStatus(
				context.TODO(),
				orderID,
				PaymentStatus_PAID,
			); err != nil {
				slog.Error("updated status", "err", err, "orderId", orderID)
				continue
			}

			// TODO publish update
		}
	}()
	return messages
}

// prepareNewOrder validates order dependencies and calculates totals. then build the complete order
// for database insertion.
func (x *OrderService) prepareNewOrder(newOrder *pb.CreatePlaceOrderRequest) (*newPlaceOrder, error) {

	ctx := context.TODO()

	customer, err := x.customer.GetCustomer(ctx, &pb.GetCustomerRequest{
		CustomerId: newOrder.CustomerId,
	})
	if err != nil {
		return nil, err
	}

	merchant, err := x.merchant.GetMerchant(ctx, &pb.GetMerchantRequest{MerchantId: newOrder.MerchantId})
	if err != nil {
		return nil, err
	}

	deliveryFee, err := x.delivery.GetDeliveryFee(ctx, &pb.GetDeliveryFeeRequest{
		CustomerId:        newOrder.CustomerId,
		CustomerAddressId: newOrder.CustomerAddressId,
		MerchantId:        newOrder.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	if newOrder.CouponCode != "" {
		if _, err := x.coupon.RedeemCoupon(ctx, &pb.RedeemCouponRequest{
			Code: newOrder.CouponCode,
		}); err != nil {
			return nil, fmt.Errorf("Failed to redeem coupon: %v", err)
		}
	}

	foodCost := calcFoodCost(merchant.Menu, newOrder.Items)
	total := foodCost - newOrder.Discount

	var items []*dbOrderItem
	for _, item := range newOrder.Items {
		items = append(items, &dbOrderItem{
			ItemID:   item.ItemId,
			Quantity: item.Quantity,
			Note:     item.Note,
		})
	}

	var selectedAddr *pb.Address
	for _, addr := range customer.Addresses {
		if addr.AddressId == newOrder.CustomerAddressId {
			selectedAddr = addr
		}
	}

	return &newPlaceOrder{
		RequestID:       newOrder.RequestId,
		CustomerID:      newOrder.CustomerId,
		MerchantID:      newOrder.MerchantId,
		Items:           items,
		CouponCode:      newOrder.CouponCode,
		Discount:        newOrder.Discount,
		DeliveryFee:     deliveryFee.Fee,
		Total:           total,
		CustomerAddress: toDbAddress(selectedAddr),
		MerchantAddress: toDbAddress(merchant.Address),
		CustomerPhone:   customer.Phone,
		PaymentMethods:  dbPaymentMethods(newOrder.PaymentMethods),
	}, nil
}

func calcFoodCost(menu []*pb.MenuItem, orderItems []*pb.OrderItem) int32 {
	menuPrices := make(map[string]int32)
	for _, menuItem := range menu {
		menuPrices[menuItem.ItemId] = menuItem.Price
	}

	var foodCost int32
	for _, newItem := range orderItems {
		if price, ok := menuPrices[newItem.ItemId]; ok {
			foodCost += price * newItem.Quantity
		}
	}
	return foodCost
}

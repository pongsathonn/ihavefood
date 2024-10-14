package internal

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

type OrderRepository interface {
	SaveNewPlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*SaveNewPlaceOrderResponse, error)
	PlaceOrders(ctx context.Context, username string) (*pb.ListUserPlaceOrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderNO string, status pb.OrderStatus) error
	UpdatePaymentStatus(ctx context.Context, orderNo string, status pb.PaymentStatus) error
	IsDuplicatedOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (bool, error)
}

type SaveNewPlaceOrderResponse struct {
	OrderNo       string
	PaymentStatus pb.PaymentStatus
	OrderStatus   pb.OrderStatus
	Created_at    int64
}

type orderRepository struct {
	client *mongo.Client
}

func NewOrderRepository(client *mongo.Client) OrderRepository {
	return &orderRepository{client: client}
}

// SaveNewPlaceOrder inserts a new order into the database.
func (r *orderRepository) SaveNewPlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*SaveNewPlaceOrderResponse, error) {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	placeOrder := preparePlaceOrder(in)

	res, err := coll.InsertOne(ctx, placeOrder)
	if err != nil {
		return nil, err
	}

	orderNo, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("failed to convert order number to primitive.ObjectId")
	}

	slog.Info("inserted new order",
		"orderNo", orderNo.Hex(),
		"createdAt", placeOrder.OrderTimeStamps.CreatedAt,
	)

	return &SaveNewPlaceOrderResponse{
		OrderNo:       orderNo.Hex(),
		PaymentStatus: pb.PaymentStatus(placeOrder.PaymentStatus),
		OrderStatus:   pb.OrderStatus(placeOrder.OrderStatus),
		Created_at:    placeOrder.OrderTimeStamps.CompletedAt,
	}, nil

}

// list all user's placeorder by username ( like query placeorder history )
func (r *orderRepository) PlaceOrders(ctx context.Context, username string) (*pb.ListUserPlaceOrderResponse, error) {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	filter := bson.D{{"username", username}}
	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var placeOrders []*pb.PlaceOrder

	for cur.Next(ctx) {
		var res PlaceOrderEntity
		if err := cur.Decode(&res); err != nil {
			return nil, err
		}
		placeOrder := entityToProtoMessage(&res)
		placeOrders = append(placeOrders, placeOrder)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return &pb.ListUserPlaceOrderResponse{PlaceOrders: placeOrders}, nil
}

// UpdateOrderStatus updates the status of a placed order.
// Available statuses are:
// - "PREPARING_ORDER"
// - "FINDING_RIDER"
// - "ONGOING"
// - "DELIVERED"
//
// Updating to "PENDING" will result in an error, as it is the default status.
// Updating to "CANCELLED" is not allowed; use this status when deleting a placed order.
func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderNo string, status pb.OrderStatus) error {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	orderNumber, err := primitive.ObjectIDFromHex(orderNo)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	var timeStampField string

	switch status {
	case pb.OrderStatus_PENDING:
		return errors.New("pending is default status")
	case pb.OrderStatus_CANCELLED:
		return errors.New("cannot update an order status to cancel")
	case pb.OrderStatus_DELIVERED:
		timeStampField = "completedAt"
	default:
		timeStampField = "updatedAt"
	}

	filter := bson.M{"_id": orderNumber}
	update := bson.M{
		"orderStatus": status,
		"orderTimeStamps": bson.M{
			timeStampField: now,
		},
	}

	if err := coll.FindOneAndUpdate(ctx, filter, update).Err(); err != nil {
		return err
	}

	slog.Info("updated order status",
		"orderNo", orderNo,
		"newStatus", status.String(),
		timeStampField, now,
	)

	return nil
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, orderNo string, status pb.PaymentStatus) error {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	orderNumber, err := primitive.ObjectIDFromHex(orderNo)
	if err != nil {
		return err
	}

	if status == pb.PaymentStatus_UNPAID {
		return errors.New("unpaid is default status")
	}

	filter := bson.M{"_id": orderNumber}
	update := bson.M{"paymentStatus": status}

	if err := coll.FindOneAndUpdate(ctx, filter, update).Err(); err != nil {
		return err
	}

	slog.Info("updated payment status",
		"orderNo", orderNo,
		"newStatus", status.String(),
	)

	return nil
}

// IsDuplicatedOrder prevents placing a duplicate order with same restaurant
// An order is considered a duplicate if:
// - The payment status is "unpaid".
// - The order was created within the last 30 minutes.
func (r *orderRepository) IsDuplicatedOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (bool, error) {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	halfHourAgo := time.Now().Add(-30 * time.Minute).Unix()

	filter := bson.M{
		"restaurantNo":              in.RestaurantNo,
		"username":                  in.Username,
		"paymentStatus":             int32(pb.PaymentStatus_UNPAID),
		"orderTimeStamps.createdAt": bson.M{"$gte": halfHourAgo},
	}

	if err := coll.FindOne(ctx, filter).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func preparePlaceOrder(in *pb.HandlePlaceOrderRequest) *PlaceOrderEntity {

	var menus []*MenuEntity
	for _, m := range in.Menus {
		menus = append(menus, &MenuEntity{
			FoodName: m.FoodName,
			Price:    m.Price,
		})
	}
	return &PlaceOrderEntity{
		// Order Number map with mongo _id (auto generate)
		RestaurantNo:   in.RestaurantNo,
		Username:       in.Username,
		CouponCode:     in.CouponCode,
		CouponDiscount: in.CouponDiscount,
		Menus:          menus,
		DeliveryFee:    in.DeliveryFee,
		Total:          in.Total,
		UserAddress: &AddressEntity{
			AddressName: in.UserAddress.AddressName,
			SubDistrict: in.UserAddress.SubDistrict,
			District:    in.UserAddress.District,
			Province:    in.UserAddress.Province,
			PostalCode:  in.UserAddress.Province,
		},
		RestaurantAddress: &AddressEntity{
			AddressName: in.UserAddress.AddressName,
			SubDistrict: in.UserAddress.SubDistrict,
			District:    in.UserAddress.District,
			Province:    in.UserAddress.Province,
			PostalCode:  in.UserAddress.Province,
		},
		UserContact: &ContactInfoEntity{
			PhoneNumber: in.UserContact.PhoneNumber,
			Email:       in.UserContact.Email,
		},
		PaymentMethods: PaymentMethodsEntity(in.PaymentMethods),
		PaymentStatus:  PaymentStatusEntity(pb.PaymentStatus_UNPAID),
		OrderStatus:    OrderStatusEntity(pb.OrderStatus_PENDING),
		OrderTimeStamps: &OrderTimestampsEntity{
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   0,
			CompletedAt: 0,
		},
	}
}

func entityToProtoMessage(entity *PlaceOrderEntity) *pb.PlaceOrder {
	var menus []*pb.Menu
	for _, m := range entity.Menus {
		menus = append(menus, &pb.Menu{
			FoodName: m.FoodName,
			Price:    m.Price,
		})

	}

	return &pb.PlaceOrder{
		No:             entity.OrderNo.Hex(),
		Username:       entity.Username,
		RestaurantNo:   entity.RestaurantNo,
		Menus:          menus,
		CouponCode:     entity.CouponCode,
		CouponDiscount: entity.CouponDiscount,
		DeliveryFee:    entity.DeliveryFee,
		Total:          entity.Total,
		UserAddress: &pb.Address{
			AddressName: entity.UserAddress.AddressName,
			SubDistrict: entity.UserAddress.SubDistrict,
			District:    entity.UserAddress.District,
			Province:    entity.UserAddress.Province,
			PostalCode:  entity.UserAddress.PostalCode,
		},
		RestaurantAddress: &pb.Address{
			AddressName: entity.RestaurantAddress.AddressName,
			SubDistrict: entity.RestaurantAddress.SubDistrict,
			District:    entity.RestaurantAddress.District,
			Province:    entity.RestaurantAddress.Province,
			PostalCode:  entity.RestaurantAddress.PostalCode,
		},
		UserContact: &pb.ContactInfo{
			PhoneNumber: entity.UserContact.PhoneNumber,
			Email:       entity.UserContact.Email,
		},
		PaymentMethods: pb.PaymentMethods(entity.PaymentMethods),
		PaymentStatus:  pb.PaymentStatus(entity.PaymentStatus),
		OrderStatus:    pb.OrderStatus(entity.OrderStatus),
		OrderTimestamps: &pb.OrderTimestamps{
			CreatedAt:   entity.OrderTimeStamps.CompletedAt,
			UpdatedAt:   entity.OrderTimeStamps.UpdatedAt,
			CompletedAt: entity.OrderTimeStamps.CompletedAt,
		},
	}

}

package internal

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

var (
	errDuplicatedOrder error = errors.New("order duplicated")
)

type SavePlaceOrderResponse struct {
	OrderId         string
	OrderTrackingId string
	PaymentStatus   pb.PaymentStatus
	OrderStatus     pb.OrderStatus
	Created_at      int64
}

type OrderRepository interface {
	PlaceOrder(ctx context.Context, username string) (*pb.ListUserPlaceOrderResponse, error)
	SavePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*SavePlaceOrderResponse, error)
}

type orderRepository struct {
	client *mongo.Client
}

func NewOrderRepository(client *mongo.Client) OrderRepository {
	return &orderRepository{client: client}
}

// TODO improve doc , adn might be change func name to plural
// list all user's placeorder by username ( like query placeorder history )
func (r *orderRepository) PlaceOrder(ctx context.Context, username string) (*pb.ListUserPlaceOrderResponse, error) {

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

	return &pb.ListUserPlaceOrderResponse{PlaceOrders: placeOrders}, nil
}

func (r *orderRepository) SavePlaceOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) (*SavePlaceOrderResponse, error) {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	if err := r.isDuplicateOrder(ctx, in); err != nil {
		return nil, err
	}

	placeOrder := preparePlaceOrder(in)

	res, err := coll.InsertOne(ctx, placeOrder)
	if err != nil {
		return nil, err
	}

	orderId, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("failed to convert id to primitive.ObjectId")
	}

	return &SavePlaceOrderResponse{
		OrderId:         orderId.Hex(),
		OrderTrackingId: placeOrder.TrackingId.Hex(),
		PaymentStatus:   pb.PaymentStatus(placeOrder.PaymentStatus),
		OrderStatus:     pb.OrderStatus(placeOrder.OrderStatus),
		Created_at:      placeOrder.OrderTimeStamps.CompletedAt,
	}, nil

}

// TODO check doc grammar
// isDuplicateOrder prevents placing a duplicate order with same restaurant
// An order is considered a duplicate if:
// - The payment status is "unpaid".
// - The order was created within the last 30 minutes.
//
// If such an order exists, the function returns the error `errDuplicatedOrder`,
// indicating the request is a duplicate. Otherwise, the order can be placed.
//
// Returns:
// - `errDuplicatedOrder` if a duplicate order is found.
// - Any other error from the query or decoding process.
//
// If User need to change or add order information then shouuld call that function
// such as AddMenus, ChangeMenus, ChangeCoupon etc.
func (r *orderRepository) isDuplicateOrder(ctx context.Context, in *pb.HandlePlaceOrderRequest) error {

	coll := r.client.Database("order_database", nil).Collection("orderCollection")

	halfHourAgo := time.Now().Add(-30 * time.Minute).Unix()

	// duplicatedOrder checks if an order with the given OrderID exists,
	// has an unpaid status, and was created within the last 30 minutes.
	duplicatedOrder, err := coll.CountDocuments(ctx, bson.M{
		"restaurantId":              in.RestaurantId,
		"username":                  in.Username,
		"paymentStatus":             int32(pb.PaymentStatus_UNPAID),
		"orderTimeStamps.createdAt": bson.M{"$gte": halfHourAgo},
	},
	)
	if err != nil {
		return err
	}

	if duplicatedOrder > 0 {
		return errDuplicatedOrder
	}

	return nil

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
		// order id map with mongo _id (auto generate)
		TrackingId:     primitive.NewObjectID(),
		RestaurantId:   in.RestaurantId,
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
		PaymentMethod: PaymentMethodEntity(in.PaymentMethod),
		PaymentStatus: PaymentStatusEntity(pb.PaymentStatus_UNPAID),
		OrderStatus:   OrderStatusEntity(pb.OrderStatus_PENDING),
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
		OrderId:         entity.OrderId.Hex(),
		OrderTrackingId: entity.TrackingId.Hex(),
		Username:        entity.Username,
		RestaurantId:    entity.RestaurantId,
		Menus:           menus,
		CouponCode:      entity.CouponCode,
		CouponDiscount:  entity.CouponDiscount,
		DeliveryFee:     entity.DeliveryFee,
		Total:           entity.Total,
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
		PaymentMethod: pb.PaymentMethod(entity.PaymentMethod),
		PaymentStatus: pb.PaymentStatus(entity.PaymentStatus),
		OrderStatus:   pb.OrderStatus(entity.OrderStatus),
		OrderTimestamps: &pb.OrderTimestamps{
			CreatedAt:   entity.OrderTimeStamps.CompletedAt,
			UpdatedAt:   entity.OrderTimeStamps.UpdatedAt,
			CompletedAt: entity.OrderTimeStamps.CompletedAt,
		},
	}

}

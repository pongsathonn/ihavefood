package internal

import (
	"time"

	// "google.golang.org/protobuf/types/known/timestamppb"
	"github.com/google/uuid"
	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
)

type newPlaceOrder struct {
	RequestID       string
	CustomerID      string
	MerchantID      string
	Items           []*dbOrderItem
	CouponCode      string
	CouponDiscount  int32
	DeliveryFee     int32
	Total           int32
	CustomerAddress *dbAddress
	MerchantAddress *dbAddress
	CustomerPhone   string
	PaymentMethods  dbPaymentMethods
}

// if not omitempty at id Mongo will use zero as id ( when insert )
type dbPlaceOrder struct {
	//OrderID           primitive.ObjectID `bson:"_id,omitempty"`
	OrderID         string           `bson:"_id,omitempty"`
	RequestID       string           `bson:"requestId"`
	CustomerID      string           `bson:"customerId"`
	MerchantID      string           `bson:"merchantId"`
	Items           []*dbOrderItem   `bson:"menus"`
	CouponCode      string           `bson:"couponCode"`
	CouponDiscount  int32            `bson:"couponDiscount"`
	DeliveryFee     int32            `bson:"deliveryFee"`
	Total           int32            `bson:"total"`
	CustomerAddress *dbAddress       `bson:"customerAddress"`
	MerchantAddress *dbAddress       `bson:"merchantAddress"`
	CustomerPhone   string           `bson:"customerPhone"`
	PaymentMethods  dbPaymentMethods `bson:"paymentMethods"`
	PaymentStatus   dbPaymentStatus  `bson:"paymentStatus"`
	OrderStatus     dbOrderStatus    `bson:"orderStatus"`
	Timestamps      *dbTimestamps    `bson:"timestamps"`
}

type dbOrderItem struct {
	ItemID   string `bson:"item_id"`
	Quantity int32  `bson:"quantity"`
	Note     string `bson:"note"`
}

type dbAddress struct {
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

type dbPaymentMethods int32

const (
	PaymentMethods_PAYMENT_METHOD_CASH        dbPaymentMethods = 0
	PaymentMethods_PAYMENT_METHOD_CREDIT_CARD dbPaymentMethods = 1
)

type dbPaymentStatus int32

const (
	PaymentStatus_UNPAID dbPaymentStatus = 0
	PaymentStatus_PAID   dbPaymentStatus = 1
)

type dbOrderStatus int32

const (
	OrderStatus_PENDING         dbOrderStatus = 0
	OrderStatus_PREPARING_ORDER dbOrderStatus = 1
	OrderStatus_FINDING_RIDER   dbOrderStatus = 2
	OrderStatus_ONGOING         dbOrderStatus = 3
	OrderStatus_DELIVERED       dbOrderStatus = 4
	OrderStatus_CANCELLED       dbOrderStatus = 5
)

type dbTimestamps struct {
	CreateTime   time.Time `bson:"createTime"`
	UpdateTime   time.Time `bson:"updateTime"`
	CompleteTime time.Time `bson:"completeTime"`
}

// EventToStatus maps event to status after an event.
var EventToStatus = map[any]pb.OrderStatus{
	pb.OrderEvent_ORDER_PLACED_EVENT:      pb.OrderStatus_PENDING,
	pb.OrderEvent_MERCHANT_ACCEPTED_EVENT: pb.OrderStatus_PREPARING_ORDER,
	pb.OrderEvent_RIDER_NOTIFIED_EVENT:    pb.OrderStatus_FINDING_RIDER,
	pb.OrderEvent_RIDER_ASSIGNED_EVENT:    pb.OrderStatus_WAIT_FOR_PICKUP,
	pb.OrderEvent_RIDER_PICKED_UP_EVENT:   pb.OrderStatus_ONGOING,
	pb.OrderEvent_RIDER_DELIVERED_EVENT:   pb.OrderStatus_DELIVERED,
	pb.OrderEvent_ORDER_CANCELLED_EVENT:   pb.OrderStatus_CANCELLED,
}

func toDbPlaceOrder(n *newPlaceOrder) *dbPlaceOrder {
	if n == nil {
		return nil
	}

	now := time.Now()
	return &dbPlaceOrder{
		OrderID:         uuid.New().String(),
		RequestID:       n.RequestID,
		CustomerID:      n.CustomerID,
		MerchantID:      n.MerchantID,
		Items:           n.Items,
		CouponCode:      n.CouponCode,
		CouponDiscount:  n.CouponDiscount,
		DeliveryFee:     n.DeliveryFee,
		Total:           n.Total,
		CustomerAddress: n.CustomerAddress,
		MerchantAddress: n.MerchantAddress,
		CustomerPhone:   n.CustomerPhone,
		PaymentMethods:  n.PaymentMethods,
		PaymentStatus:   PaymentStatus_UNPAID,
		OrderStatus:     OrderStatus_PENDING,
		Timestamps: &dbTimestamps{
			CreateTime:   now,
			UpdateTime:   now,
			CompleteTime: time.Time{},
		},
	}
}

func toDbAddress(addr *pb.Address) *dbAddress {
	if addr == nil {
		return nil
	}
	return &dbAddress{
		AddressName: addr.AddressName,
		SubDistrict: addr.SubDistrict,
		District:    addr.District,
		Province:    addr.Province,
		PostalCode:  addr.PostalCode,
	}
}

func toProtoAddress(addr *dbAddress) *pb.Address {
	if addr == nil {
		return nil
	}
	return &pb.Address{
		AddressName: addr.AddressName,
		SubDistrict: addr.SubDistrict,
		District:    addr.District,
		Province:    addr.Province,
		PostalCode:  addr.PostalCode,
	}
}

func toProtoPlaceOrder(order *dbPlaceOrder) *pb.PlaceOrder {

	if order == nil {
		return nil
	}

	var items []*pb.OrderItem
	for _, item := range order.Items {
		items = append(items, &pb.OrderItem{
			ItemId:   item.ItemID,
			Quantity: item.Quantity,
			Note:     item.Note,
		})
	}

	return &pb.PlaceOrder{
		RequestId:       order.RequestID,
		OrderId:         order.OrderID,
		CustomerId:      order.CustomerID,
		MerchantId:      order.MerchantID,
		Items:           items,
		CouponDiscount:  order.CouponDiscount,
		DeliveryFee:     order.DeliveryFee,
		Total:           order.Total,
		CustomerAddress: toProtoAddress(order.CustomerAddress),
		MerchantAddress: toProtoAddress(order.MerchantAddress),
		CustomerPhone:   order.CustomerPhone,
		CouponCode:      order.CouponCode,
		PaymentMethods:  pb.PaymentMethods(order.PaymentMethods),
		PaymentStatus:   pb.PaymentStatus(order.PaymentStatus),
		OrderStatus:     pb.OrderStatus(order.OrderStatus),
		// Timestamps: &pb.OrderEventTimestamps{
		// 	OrderPlacedTime: timestamppb.New(order.Timestamps.TODO)
		// 	// TODO:
		// },
	}
}

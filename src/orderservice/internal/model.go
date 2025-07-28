package internal

import (
	"time"
)

type newPlaceOrder struct {
	RequestID       string
	CustomerID      string
	MerchantID      string
	Menu            []*dbMenuItem
	CouponCode      string
	CouponDiscount  int32
	DeliveryFee     int32
	Total           int32
	CustomerAddress *dbAddress
	MerchantAddress *dbAddress
	CustomerContact *dbContactInfo
	PaymentMethods  dbPaymentMethods
}

// if not omitempty at id Mongo will use zero as id ( when insert )
type dbPlaceOrder struct {
	//OrderID           primitive.ObjectID `bson:"_id,omitempty"`
	OrderID         string           `bson:"_id,omitempty"`
	RequestID       string           `bson:"requestId"`
	CustomerID      string           `bson:"customerId"`
	MerchantID      string           `bson:"merchantId"`
	Menu            []*dbMenuItem    `bson:"menus"`
	CouponCode      string           `bson:"couponCode"`
	CouponDiscount  int32            `bson:"couponDiscount"`
	DeliveryFee     int32            `bson:"deliveryFee"`
	Total           int32            `bson:"total"`
	CustomerAddress *dbAddress       `bson:"customerAddress"`
	MerchantAddress *dbAddress       `bson:"merchantAddress"`
	CustomerContact *dbContactInfo   `bson:"customerContact"`
	PaymentMethods  dbPaymentMethods `bson:"paymentMethods"`
	PaymentStatus   dbPaymentStatus  `bson:"paymentStatus"`
	OrderStatus     dbOrderStatus    `bson:"orderStatus"`
	Timestamps      *dbTimestamps    `bson:"timestamps"`
}

type dbMenuItem struct {
	FoodName string `bson:"foodName"`
	Price    int32  `bson:"price"`
}

type dbAddress struct {
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

type dbContactInfo struct {
	PhoneNumber string `bson:"phoneNumber"`
	Email       string `bson:"email"`
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

package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// The term "entity" refers to either a table in the relational model,
// or a collection in the MongoDB model

// This PlaceOrderEntity use as model for "INSERT" and "QUERY" with mongodb
// to make this model match with protobuff json name should be same as bson name tag
// if not omitempty at id Mongo will use zero as id ( when insert )
type PlaceOrderEntity struct {
	OrderNo           primitive.ObjectID   `bson:"_id,omitempty"`
	Username          string               `bson:"username"`
	RestaurantNo      string               `bson:"restaurantNo"`
	Menus             []*MenuEntity        `bson:"menus"`
	CouponCode        string               `bson:"couponCode"`
	CouponDiscount    int32                `bson:"couponDiscount"`
	DeliveryFee       int32                `bson:"deliveryFee"`
	Total             int32                `bson:"total"`
	UserAddress       *AddressEntity       `bson:"userAddress"`
	RestaurantAddress *AddressEntity       `bson:"restaurantAddress"`
	UserContact       *ContactInfoEntity   `bson:"userContact"`
	PaymentMethods    PaymentMethodsEntity `bson:"paymentMethods"`
	PaymentStatus     PaymentStatusEntity  `bson:"paymentStatus"`
	OrderStatus       OrderStatusEntity    `bson:"orderStatus"`
	Timestamps        *TimestampsEntity    `bson:"timestamps"`
}

type MenuEntity struct {
	FoodName string `bson:"foodName"`
	Price    int32  `bson:"price"`
}

type AddressEntity struct {
	AddressName string `bson:"addressName"`
	SubDistrict string `bson:"subDistrict"`
	District    string `bson:"district"`
	Province    string `bson:"province"`
	PostalCode  string `bson:"postalCode"`
}

type ContactInfoEntity struct {
	PhoneNumber string `bson:"phoneNumber"`
	Email       string `bson:"email"`
}

type PaymentMethodsEntity int32

const (
	PaymentMethods_PAYMENT_METHOD_CASH        PaymentMethodsEntity = 0
	PaymentMethods_PAYMENT_METHOD_CREDIT_CARD PaymentMethodsEntity = 1
)

type PaymentStatusEntity int32

const (
	PaymentStatus_UNPAID PaymentStatusEntity = 0
	PaymentStatus_PAID   PaymentStatusEntity = 1
)

type OrderStatusEntity int32

const (
	OrderStatus_PENDING         OrderStatusEntity = 0
	OrderStatus_PREPARING_ORDER OrderStatusEntity = 1
	OrderStatus_FINDING_RIDER   OrderStatusEntity = 2
	OrderStatus_ONGOING         OrderStatusEntity = 3
	OrderStatus_DELIVERED       OrderStatusEntity = 4
	OrderStatus_CANCELLED       OrderStatusEntity = 5
)

// TODO use time.Time for readable
type TimestampsEntity struct {
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
	CompletedAt time.Time `bson:"completedAt"`
}

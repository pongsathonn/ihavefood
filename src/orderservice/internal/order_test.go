package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/pongsathonn/ihavefood/src/orderservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) ListPlaceOrders(ctx context.Context, customerID string) ([]*dbPlaceOrder, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Note: ListPlaceOrders returns []*dbPlaceOrder, the mock should handle this type.
	return args.Get(0).([]*dbPlaceOrder), args.Error(1)
}

func (m *MockStorage) GetPlaceOrder(ctx context.Context, orderID string) (*dbPlaceOrder, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dbPlaceOrder), args.Error(1)
}

func (m *MockStorage) Create(ctx context.Context, order *newPlaceOrder) (string, error) {
	args := m.Called(ctx, order)
	return args.String(0), args.Error(1)
}

// UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (string, error)
func (m *MockStorage) UpdateOrderStatus(ctx context.Context, orderID string, status dbOrderStatus) (bool, error) {
	args := m.Called(ctx, orderID, status)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) UpdatePaymentStatus(ctx context.Context, orderID string, status dbPaymentStatus) (bool, error) {
	args := m.Called(ctx, orderID, status)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) DeletePlaceOrder(ctx context.Context, orderID string) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

// Mock RabbitMQ implementation
type MockRabbitMQ struct {
	mock.Mock
}

func (m *MockRabbitMQ) Publish(ctx context.Context, routingKey string, msg amqp.Publishing) error {
	args := m.Called(ctx, routingKey, msg)
	return args.Error(0)
}

func (m *MockRabbitMQ) Subscribe(ctx context.Context, queue, key string) (<-chan amqp.Delivery, error) {
	args := m.Called(ctx, queue, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Note: This requires a mock channel, but is not tested in the main CRUD tests.
	return args.Get(0).(<-chan amqp.Delivery), args.Error(1)
}

type MockCouponClient struct {
	mock.Mock
}

func (m *MockCouponClient) GetCoupon(ctx context.Context, in *pb.GetCouponRequest, opts ...grpc.CallOption) (*pb.Coupon, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Coupon), args.Error(1)
}

type MockCustomerClient struct {
	mock.Mock
}

func (m *MockCustomerClient) GetCustomer(ctx context.Context, in *pb.GetCustomerRequest, opts ...grpc.CallOption) (*pb.Customer, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Customer), args.Error(1)
}

type MockDeliveryClient struct {
	mock.Mock
}

func (m *MockDeliveryClient) GetDeliveryFee(ctx context.Context, in *pb.GetDeliveryFeeRequest, opts ...grpc.CallOption) (*pb.GetDeliveryFeeResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetDeliveryFeeResponse), args.Error(1)
}

type MockMerchantClient struct {
	mock.Mock
}

func (m *MockMerchantClient) GetMerchant(ctx context.Context, in *pb.GetMerchantRequest, opts ...grpc.CallOption) (*pb.Merchant, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Merchant), args.Error(1)
}

// =================================================================================
// 									TEST
// =================================================================================

// Test: Successful Order Creation
func TestCreatePlaceOrder_Success(t *testing.T) {
	// Setup validator
	SetupValidator()

	// Setup mocks
	mockStorage := new(MockStorage)
	mockRabbitMQ := new(MockRabbitMQ)
	mockCoupon := new(MockCouponClient)
	mockCustomer := new(MockCustomerClient)
	mockDelivery := new(MockDeliveryClient)
	mockMerchant := new(MockMerchantClient)

	orderService := NewOrderService(
		mockStorage,
		mockRabbitMQ,
		mockCoupon,
		mockCustomer,
		mockDelivery,
		mockMerchant,
	)

	// Test request
	req := &pb.CreatePlaceOrderRequest{
		RequestId:         "10000000-0000-4000-8000-000000000001",
		CustomerId:        "20000000-0000-4000-8000-000000000002",
		MerchantId:        "30000000-0000-4000-8000-000000000003",
		CustomerAddressId: "40000000-0000-4000-8000-000000000004",
		CouponCode:        "DISCOUNT10",
		PaymentMethods:    pb.PaymentMethods_PAYMENT_METHOD_CREDIT_CARD,
		Items: []*pb.OrderItem{
			{ItemId: "50000000-0000-4000-8000-000000000005", Quantity: 2, Note: "no spice"},
		},
	}

	expectedOrderID := "11100020-0000-4000-8000-000000000001"

	// Mocked values
	menuPrice := int32(100)
	quantity := req.Items[0].Quantity
	deliveryFee := int32(50)
	couponDiscount := int32(20)

	expectedTotal := menuPrice*int32(quantity) + deliveryFee - couponDiscount

	mockCustomer.On("GetCustomer", mock.Anything, mock.AnythingOfType("*genproto.GetCustomerRequest")).Return(&pb.Customer{
		CustomerId: req.CustomerId,
		Phone:      "0812345678",
		Addresses: []*pb.Address{
			{AddressId: req.CustomerAddressId, AddressName: "123 Street"},
		},
	}, nil)

	mockMerchant.On("GetMerchant", mock.Anything, mock.AnythingOfType("*genproto.GetMerchantRequest")).Return(&pb.Merchant{
		MerchantId: req.MerchantId,
		Address:    &pb.Address{AddressName: "Merch Address"},
		Menu: []*pb.MenuItem{
			{ItemId: req.Items[0].ItemId, Price: menuPrice},
		},
	}, nil)

	mockDelivery.On("GetDeliveryFee", mock.Anything, mock.AnythingOfType("*genproto.GetDeliveryFeeRequest")).Return(&pb.GetDeliveryFeeResponse{
		Fee: deliveryFee,
	}, nil)

	mockCoupon.On("GetCoupon", mock.Anything, mock.AnythingOfType("*genproto.GetCouponRequest")).Return(&pb.Coupon{
		Code:     req.CouponCode,
		Discount: couponDiscount,
	}, nil)

	mockStorage.On("Create", mock.Anything, mock.AnythingOfType("*internal.newPlaceOrder")).Return(expectedOrderID, nil)
	mockStorage.On("GetPlaceOrder", mock.Anything, expectedOrderID).Return(&dbPlaceOrder{
		OrderID:    expectedOrderID,
		CustomerID: req.CustomerId,
		MerchantID: req.MerchantId,
		Total:      expectedTotal,
	}, nil)

	mockRabbitMQ.On("Publish", mock.Anything, "order.placed.event", mock.MatchedBy(func(p amqp.Publishing) bool {
		if p.Type != "ihavefood.PlaceOrder" {
			return false
		}
		var publishedOrder pb.PlaceOrder
		if err := proto.Unmarshal(p.Body, &publishedOrder); err != nil {
			t.Errorf("Failed to unmarshal published message body: %v", err)
			return false
		}
		return publishedOrder.OrderId == expectedOrderID
	})).Return(nil)

	// Execute
	result, err := orderService.CreatePlaceOrder(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedOrderID, result.OrderId)
	assert.Equal(t, expectedTotal, result.Total)
}

// Test: Validation Failure
func TestCreatePlaceOrder_ValidationError(t *testing.T) {
	// Setup validator
	SetupValidator()
	orderService := &OrderService{}

	req := &pb.CreatePlaceOrderRequest{
		// Missing required fields
		CustomerId:     "invalid-uuid",                               // Fails UUID4 check
		MerchantId:     "",                                           // Fails required check
		Items:          []*pb.OrderItem{},                            // Fails vitems check (length < 1)
		PaymentMethods: pb.PaymentMethods_PAYMENT_METHOD_UNSPECIFIED, // Fails vpayment_method
	}

	result, err := orderService.CreatePlaceOrder(context.Background(), req)

	assert.Nil(t, result)
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

}

// TestCreatePlaceOrder_Failure_DeletesOrder ensures a created order is deleted if subsequent operations fail.
func TestCreatePlaceOrder_Failure_DeletesOrder(t *testing.T) {

	SetupValidator()

	mockStorage := new(MockStorage)
	mockRabbitMQ := new(MockRabbitMQ)
	mockCustomer := new(MockCustomerClient)
	mockMerchant := new(MockMerchantClient)
	mockDelivery := new(MockDeliveryClient)
	mockCoupon := new(MockCouponClient)

	// Mock successful client calls for prepareNewOrder
	mockCustomer.On("GetCustomer", mock.Anything, mock.Anything).Return(&pb.Customer{
		CustomerId: "cust-123",
		Phone:      "0812345678",
		Addresses:  []*pb.Address{{AddressId: "addr-123"}},
	}, nil)
	mockMerchant.On("GetMerchant", mock.Anything, mock.Anything).Return(&pb.Merchant{
		MerchantId: "merch-123",
		Address:    &pb.Address{},
		Menu:       []*pb.MenuItem{{ItemId: "50000000-0000-4000-8000-000000000005", Price: 100}},
	}, nil)
	mockDelivery.On("GetDeliveryFee", mock.Anything, mock.Anything).Return(&pb.GetDeliveryFeeResponse{Fee: 50}, nil)
	mockCoupon.On("GetCoupon", mock.Anything, mock.Anything).Return(&pb.Coupon{Discount: 0}, nil)

	// --------------
	newOrderID := "10111000-0000-4000-8000-000000000005"
	mockStorage.On("Create", mock.Anything, mock.Anything).Return(newOrderID, nil)
	mockStorage.On("GetPlaceOrder", mock.Anything, newOrderID).Return(&dbPlaceOrder{}, nil)
	mockRabbitMQ.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("publish failed"))
	mockStorage.On("DeletePlaceOrder", mock.Anything, newOrderID).Return(nil)

	orderService := NewOrderService(
		mockStorage,
		mockRabbitMQ,
		mockCoupon,
		mockCustomer,
		mockDelivery,
		mockMerchant,
	)

	req := &pb.CreatePlaceOrderRequest{
		RequestId:         "10000000-0000-4000-8000-000000000001",
		CustomerId:        "20000000-0000-4000-8000-000000000002",
		MerchantId:        "30000000-0000-4000-8000-000000000003",
		CustomerAddressId: "40000000-0000-4000-8000-000000000004",
		Items:             []*pb.OrderItem{{ItemId: "50000000-0000-4000-8000-000000000005", Quantity: 1}},
		PaymentMethods:    pb.PaymentMethods_PAYMENT_METHOD_CASH,
	}

	result, err := orderService.CreatePlaceOrder(context.Background(), req)

	assert.Nil(t, result)
	assert.Error(t, err)
	mockStorage.AssertCalled(t, "DeletePlaceOrder", mock.Anything, newOrderID)
}

package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/customerservice/genproto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// CustomerService manages user customer.
type CustomerService struct {
	pb.UnimplementedCustomerServiceServer

	rabbitmq *RabbitMQ
	store    *customerStorage
}

func NewCustomerService(rabbitmq *RabbitMQ, store *customerStorage) *CustomerService {
	return &CustomerService{
		rabbitmq: rabbitmq,
		store:    store,
	}
}

func (x *CustomerService) ListCustomers(ctx context.Context, in *pb.ListCustomersRequest) (*pb.ListCustomersResponse, error) {

	results, err := x.store.listCustomers(ctx)
	if err != nil {
		slog.Error("storage list customers", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	var customers []*pb.Customer
	for _, customer := range results {
		customers = append(customers, dbToProto(customer))
	}

	return &pb.ListCustomersResponse{Customers: customers}, nil
}

func (x *CustomerService) GetCustomer(ctx context.Context, in *pb.GetCustomerRequest) (*pb.Customer, error) {

	customer, err := x.store.getCustomer(ctx, in.CustomerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		slog.Error("store get customer", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}
	return dbToProto(customer), nil
}

func (x *CustomerService) CreateAddress(ctx context.Context, in *pb.CreateAddressRequest) (*pb.Address, error) {

	customer, err := x.store.getCustomer(ctx, in.CustomerId)
	if err != nil {
		slog.Error("store get customer", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	maxAddr := 5
	if len(customer.Addresses) >= maxAddr {
		return nil, status.Errorf(codes.ResourceExhausted, "customer has reached the limit of %d addresses", maxAddr)
	}

	addressID, err := x.store.createAddress(ctx, in.CustomerId, &dbAddress{
		AddressName: &in.Address.AddressName,
		SubDistrict: &in.Address.SubDistrict,
		District:    &in.Address.District,
		Province:    &in.Address.Province,
		PostalCode:  &in.Address.PostalCode,
	})
	if err != nil {
		slog.Error("store create address", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	addr, err := x.store.getAddress(ctx, in.CustomerId, addressID)
	if err != nil {
		slog.Error("store get address", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.Address{
		AddressId:   addr.AddressID,
		AddressName: *addr.AddressName,
		SubDistrict: *addr.SubDistrict,
		District:    *addr.District,
		Province:    *addr.Province,
		PostalCode:  *addr.PostalCode,
	}, nil

}

func (x *CustomerService) UpdateCustomer(ctx context.Context, in *pb.UpdateCustomerRequest) (*pb.Customer, error) {

	update := &dbCustomer{
		Username: in.NewUsername,
		Social: dbSocial{
			Facebook:  &in.NewSocial.Facebook,
			Instagram: &in.NewSocial.Instagram,
			Line:      &in.NewSocial.Line,
		},
	}

	customerID, err := x.store.update(ctx, in.CustomerId, update)
	if err != nil {
		slog.Error("store update customer", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	customer, err := x.store.getCustomer(ctx, customerID)
	if err != nil {
		slog.Error("store get customer", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return dbToProto(customer), nil

}

func (x *CustomerService) DeleteCustomer(ctx context.Context, in *pb.DeleteCustomerRequest) (*emptypb.Empty, error) {

	//TODO validate intput

	err := x.store.remove(ctx, in.CustomerId)
	if err != nil {
		slog.Error("store remove customer", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &emptypb.Empty{}, nil
}

func (x *CustomerService) HandleCustomerCreation(msg amqp.Delivery) error {

	var newCustomer pb.SyncCustomerCreated
	if err := proto.Unmarshal(msg.Body, &newCustomer); err != nil {
		return err
	}

	parsed, err := uuid.Parse(newCustomer.CustomerId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return err
	}

	customerId := parsed.String()
	defaultUsername := fmt.Sprintf("customer%s", customerId[len(customerId)-4:])

	customerID, err := x.store.create(context.TODO(), &dbNewCustomer{
		CustomerID: customerId,
		Username:   defaultUsername,
		Email:      newCustomer.Email,
		CreateTime: newCustomer.CreateTime.AsTime(),
	})
	if err != nil {
		return err
	}

	slog.Info("created a new customer", "customerID", customerID)
	return nil
}

func protoToDb(customer *pb.Customer) *dbCustomer {

	var addresses []*dbAddress
	for _, a := range customer.Addresses {
		addresses = append(addresses, &dbAddress{
			AddressID:   a.AddressId,
			AddressName: &a.AddressName,
			SubDistrict: &a.SubDistrict,
			District:    &a.District,
			Province:    &a.Province,
			PostalCode:  &a.PostalCode,
		})
	}

	return &dbCustomer{
		CustomerID: customer.CustomerId,
		Username:   customer.Username,
		Email:      customer.Email,
		Phone:      customer.Phone,
		Social: dbSocial{
			Facebook:  &customer.Social.Facebook,
			Instagram: &customer.Social.Instagram,
			Line:      &customer.Social.Line,
		},
		Addresses:  addresses,
		CreateTime: customer.CreateTime.AsTime(),
		UpdateTime: customer.UpdateTime.AsTime(),
	}

}

func dbToProto(customer *dbCustomer) *pb.Customer {

	var addresses []*pb.Address
	for _, a := range customer.Addresses {
		addresses = append(addresses, &pb.Address{
			AddressId:   a.AddressID,
			AddressName: safeDeref(a.AddressName),
			SubDistrict: safeDeref(a.SubDistrict),
			District:    safeDeref(a.District),
			Province:    safeDeref(a.Province),
			PostalCode:  safeDeref(a.PostalCode),
		})
	}

	return &pb.Customer{
		CustomerId: customer.CustomerID,
		Username:   customer.Username,
		Email:      customer.Email,
		Phone:      customer.Phone,
		Addresses:  addresses,
		Social: &pb.Social{
			Facebook:  safeDeref(customer.Social.Facebook),
			Instagram: safeDeref(customer.Social.Instagram),
			Line:      safeDeref(customer.Social.Line),
		},
		CreateTime: timestamppb.New(customer.CreateTime),
	}
}

func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

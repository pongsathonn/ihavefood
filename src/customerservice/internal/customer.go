package internal

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/pongsathonn/ihavefood/src/customerservice/genproto"
)

// CustomerService manages user customer.
type CustomerService struct {
	pb.UnimplementedCustomerServiceServer

	rabbitmq *rabbitMQ
	store    *customerStorage
}

func NewCustomerService(rabbitmq *rabbitMQ, store *customerStorage) *CustomerService {
	return &CustomerService{
		rabbitmq: rabbitmq,
		store:    store,
	}
}

func (x *CustomerService) ListCustomers(ctx context.Context, in *pb.ListCustomersRequest) (*pb.ListCustomersResponse, error) {

	// TODO validate input

	results, err := x.store.listCustomers(ctx)
	if err != nil {
		return nil, err
	}

	var customers []*pb.Customer
	for _, customer := range results {
		customers = append(customers, dbToProto(customer))
	}

	return &pb.ListCustomersResponse{Customers: customers}, nil
}

func (x *CustomerService) GetCustomer(ctx context.Context, in *pb.GetCustomerRequest) (*pb.Customer, error) {

	//TODO validate

	customer, err := x.store.getCustomer(ctx, in.CustomerId)
	if err != nil {
		return nil, err
	}

	return dbToProto(customer), nil
}

func (x *CustomerService) CreateCustomer(ctx context.Context, in *pb.CreateCustomerRequest) (*pb.Customer, error) {

	uuid, err := uuid.Parse(in.CustomerId)
	if err != nil {
		slog.Error("invalid uuid", "err", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid customer id")
	}

	customerID, err := x.store.create(ctx, &newCustomer{
		CustomerID: uuid.String(),
		Username:   in.Username,
	})
	if err != nil {
		slog.Error("failed to create user customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to create user customer")
	}

	customer, err := x.store.getCustomer(ctx, customerID)
	if err != nil {
		slog.Error("failed to retrive user customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive user customer")
	}

	return dbToProto(customer), nil
}

func (x *CustomerService) CreateAddress(ctx context.Context, in *pb.CreateAddressRequest) (*pb.Customer, error) {

	// TODO validate input

	numAddr, err := x.store.countAddress(ctx, in.CustomerId)
	if err != nil {
		slog.Error("failed to count user customer address", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to count adress")
	}

	if numAddr >= 5 {
		return nil, status.Errorf(codes.ResourceExhausted, "user has reached the limit of five addresses")
	}

	customerID, err := x.store.createAddress(ctx, in.CustomerId, &dbAddress{
		AddressName: sql.NullString{String: in.Address.AddressName},
		SubDistrict: sql.NullString{String: in.Address.SubDistrict},
		District:    sql.NullString{String: in.Address.District},
		Province:    sql.NullString{String: in.Address.Province},
		PostalCode:  sql.NullString{String: in.Address.PostalCode},
	})
	if err != nil {
		slog.Error("failed to update customer address", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update adress")
	}

	// edge case: if customerID not exists it will return nil customer as nil
	customer, err := x.store.getCustomer(ctx, customerID)
	if err != nil {
		slog.Error("failed to retrive customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive customer")
	}

	return dbToProto(customer), nil
}

func (x *CustomerService) UpdateCustomer(ctx context.Context, in *pb.UpdateCustomerRequest) (*pb.Customer, error) {

	// TODO validate input

	update := &dbCustomer{
		Username: in.NewUsername,
		Bio:      sql.NullString{String: in.NewBio},
		Social: dbSocial{
			Facebook:  sql.NullString{String: in.NewSocial.Facebook},
			Instagram: sql.NullString{String: in.NewSocial.Instagram},
			Line:      sql.NullString{String: in.NewSocial.Line},
		},
	}

	customerID, err := x.store.update(ctx, in.CustomerId, update)
	if err != nil {
		slog.Error("failed to update customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update customer")
	}

	customer, err := x.store.getCustomer(ctx, customerID)
	if err != nil {
		slog.Error("failed to retrive customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to retrive customer")
	}

	return dbToProto(customer), nil

}

func (x *CustomerService) DeleteCustomer(ctx context.Context, in *pb.DeleteCustomerRequest) (*emptypb.Empty, error) {

	//TODO validate intput

	err := x.store.remove(ctx, in.CustomerId)
	if err != nil {
		slog.Error("delete user customer", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	return &emptypb.Empty{}, nil
}

func protoToDb(customer *pb.Customer) *dbCustomer {

	var addresses []*dbAddress
	for _, a := range customer.Addresses {
		addresses = append(addresses, &dbAddress{
			AddressName: sql.NullString{String: a.AddressName},
			SubDistrict: sql.NullString{String: a.SubDistrict},
			District:    sql.NullString{String: a.District},
			Province:    sql.NullString{String: a.Province},
			PostalCode:  sql.NullString{String: a.PostalCode},
		})
	}

	return &dbCustomer{
		//UserID: "",
		Username: customer.Username,
		Bio:      sql.NullString{String: customer.Bio},
		Social: dbSocial{
			Facebook:  sql.NullString{String: customer.Social.Facebook},
			Instagram: sql.NullString{String: customer.Social.Instagram},
			Line:      sql.NullString{String: customer.Social.Line},
		},
		Addresses: addresses,
		// CreateTime: nil,
	}

}

func dbToProto(customer *dbCustomer) *pb.Customer {

	var addresses []*pb.Address
	for _, a := range customer.Addresses {
		addresses = append(addresses, &pb.Address{
			AddressName: a.AddressName.String,
			SubDistrict: a.SubDistrict.String,
			District:    a.District.String,
			Province:    a.Province.String,
			PostalCode:  a.PostalCode.String,
		})
	}

	return &pb.Customer{
		CustomerId: customer.CustomerID,
		Username:   customer.Username,
		Bio:        customer.Bio.String,
		Social: &pb.Social{
			Facebook:  customer.Social.Facebook.String,
			Instagram: customer.Social.Instagram.String,
			Line:      customer.Social.Line.String,
		},
		Addresses:  addresses,
		CreateTime: timestamppb.New(customer.CreateTime),
	}
}

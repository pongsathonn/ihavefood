package main

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pongsathonn/ihavefood/src/userservice/genproto"
)

type userService struct {
	pb.UnimplementedUserServiceServer

	db *sql.DB
}

func NewUserService(db *sql.DB) *userService {
	return &userService{db: db}
}

func (x *userService) UpdateUser(context.Context, *pb.Empty) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}

func (x *userService) CreateUser(context.Context, *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUser not implemented")
}

func (x *userService) ListUser(context.Context, *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
}

func (x *userService) GetUser(context.Context, *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}

func (x *userService) DeleteUser(context.Context, *pb.DeleteUserRequest) (*pb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}

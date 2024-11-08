package internal

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errNoUsername         = status.Error(codes.InvalidArgument, "username must be provided")
	errNoPassword         = status.Error(codes.InvalidArgument, "password must be provided")
	errNoEmail            = status.Error(codes.InvalidArgument, "email must be provided")
	errNoUsernamePassword = status.Error(codes.InvalidArgument, "username or password must be provided")

	errUserIncorrect   = status.Error(codes.InvalidArgument, "username or password incorrect")
	errUserNotFound    = status.Error(codes.NotFound, "user not found")
	errPasswordHashing = status.Error(codes.Internal, "password hashing failed")

	errNoToken       = status.Error(codes.InvalidArgument, "token must be provided")
	errInvalidToken  = status.Error(codes.Unauthenticated, "invalid token")
	errGenerateToken = status.Error(codes.Internal, "failed to generate authentication token")
)

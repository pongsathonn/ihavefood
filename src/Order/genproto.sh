#!/bin/bash -eu

# ensure that you've set path for protoc
PATH=$PATH:$GOPATH/bin

# TODO: user absolute path for this 
protodir=../../protos # .proto file

# --go_out= Destination directory
# -I = where protoc search for imports

protoc -I $protodir --go_out=. --go-grpc_out=. $protodir/food.proto

       

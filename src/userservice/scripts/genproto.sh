#!/bin/bash -eu

<<comment
 -I = where protoc search for imports
 --go_out is  Destination directory
 --go-grpc_out is for complie gRPC stub
 --grpc-gateway_out is for complie gRPC-gateway 

    **for gRPC-gateway copy these files from googleapis
        google/api/annotations.proto
        google/api/field_behavior.proto
        google/api/http.proto
        google/api/httpbody.proto
comment

# ensure that you've set path for protoc
PATH=$PATH:$GOPATH/bin

# path to .proto file
protodir=../../../protos 

protoc -I $protodir --go_out=../. --go-grpc_out=../. --grpc-gateway_out=../. $protodir/food.proto

       


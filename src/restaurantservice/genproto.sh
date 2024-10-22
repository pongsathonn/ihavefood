# ensure that you've set path for protoc
PATH=$PATH:$GOPATH/bin

protodir=../../protos # .proto file

protoc -I $protodir --go_out=. --go-grpc_out=. --grpc-gateway_out=. $protodir/food.proto


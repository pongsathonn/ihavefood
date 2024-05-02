
#!/bin/bash -eu

PATH=$PATH:$GOPATH/bin

protodir=../../protos # .proto file

# --go_out= Destination directory
# -I = where proto complier search for imports

protoc -I $protodir --go_out=. --go-grpc_out=. $protodir/food.proto

       

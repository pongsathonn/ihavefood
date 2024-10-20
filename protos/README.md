



flags

    -I is where protoc search for imports
    --go_out generates .pb.go  (protobuf messages)
    --go-grpc_out generates .grpc.pb.go (gRPC service definitions)
    --grpc-gateway_out generates .gw.go (gRPC-Gateway for REST support)


more see https://protobuf.dev/reference/go/go-generated

 for gRPC-gateway copy these files from googleapis to source code
 and run go mod tidy to resolve

    google/api/annotations.proto
    google/api/field_behavior.proto
    google/api/http.proto
    google/api/httpbody.proto


a
sdoasd



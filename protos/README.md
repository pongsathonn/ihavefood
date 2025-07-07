Protoc Flags

    -I is where protoc search for imports
    --go_out generates .pb.go  (protobuf messages)
    --go-grpc_out generates .grpc.pb.go (gRPC service definitions)
    --grpc-gateway_out generates .gw.go (gRPC-Gateway for REST support)

more https://protobuf.dev/reference/go/go-generated

for gRPC-gateway copy these files from googleapis to source code
and run go mod tidy to resolve import

    google/api/annotations.proto
    google/api/field_behavior.proto
    google/api/http.proto
    google/api/httpbody.proto

more https://github.com/grpc-ecosystem/grpc-gateway#usage

dir /google

### Install protobuff complier
<pre>
 $ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
 $ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
</pre>


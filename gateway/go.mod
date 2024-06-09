module github.com/pongsathonn/food-delivery/gateway

go 1.22.2

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.34.1
)

require (
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
)

replace github.com/pongsathonn/food-delivery/src/user => ../src/user

replace github.com/pongsathonn/food-delivery/src/order => ../src/order

replace github.com/pongsathonn/food-delivery/src/coupon => ../src/coupon

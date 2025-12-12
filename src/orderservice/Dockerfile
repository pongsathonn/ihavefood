# syntax=docker/dockerfile:1
FROM  golang:1.24 AS builder
WORKDIR /src/order
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags=all="-N -l" -o /order

FROM alpine:latest
COPY --from=builder /order /order
CMD ["/order"]



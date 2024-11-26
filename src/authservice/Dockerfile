# syntax=docker/dockerfile:1

# stage 1: build
FROM golang:1.22 AS builder
WORKDIR /src/auth
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go install github.com/go-delve/delve/cmd/dlv@latest

# CGO_ENABLED=0 is build as staticaly-linked binary
# default is 1 it will rely on some C lib

# GOOS=linux is target os which is linux is this case
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth

FROM alpine:latest

# libc6 makes alpine compatible with dlv
RUN apk add --no-cache libc6-compat

COPY --from=builder /auth /auth
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv

CMD ["/auth"]



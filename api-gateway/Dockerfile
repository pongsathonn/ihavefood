# syntax=docker/dockerfile:1
FROM golang:1.23 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go build -o /gateway

FROM alpine:latest
RUN apk add --no-cache libc6-compat
COPY --from=builder /gateway /gateway
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv
CMD ["/gateway"]


# syntax=docker/dockerfile:1

FROM golang:1.23 AS builder
WORKDIR /src/customer
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags=all="-N -l" -o /customer

FROM alpine:latest

RUN apk add --no-cache libc6-compat

COPY --from=builder /customer /customer
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv

CMD ["/customer"]


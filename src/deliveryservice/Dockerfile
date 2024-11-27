# syntax=docker/dockerfile:1

# stage 1: build
FROM golang:1.22 AS builder
WORKDIR /src/delivery
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN CGO_ENABLED=0 GOOS=linux go build -o /delivery

FROM alpine:latest

RUN apk add --no-cache libc6-compat

COPY --from=builder /delivery /delivery
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv

CMD ["/delivery"]




# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /src
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=0 GOOS=linux go build -o /coupon

# Stage 2: Runtime
FROM alpine:latest

RUN apk add --no-cache libc6-compat

COPY --from=builder /coupon /coupon
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv

CMD ["/coupon"]


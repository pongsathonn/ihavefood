# syntax=docker/dockerfile:1
FROM  golang:1.24 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags=all="-N -l" -o /gateway

FROM alpine:latest
COPY --from=builder /gateway /gateway
CMD ["/gateway"]

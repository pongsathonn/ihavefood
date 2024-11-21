# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR /gateway

# restore dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy

RUN go install github.com/go-delve/delve/cmd/dlv@latest

# RUN apt update 

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /gateway

# Stage 2: Runtime
FROM alpine:latest

# uncomment to install dependencies for runtime 
# RUN apk add --no-cache ca-certificates

COPY --from=builder /gateway /gateway

# for development stage(debugging)
COPY --from=builder /go/bin/dlv /go/bin/dlv
ENV PATH=$PATH:/go/bin


CMD ["/gateway"]



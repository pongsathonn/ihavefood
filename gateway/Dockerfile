# syntax=docker/dockerfile:1
FROM golang:1.22
WORKDIR gateway

# restore dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN apt update 

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /gatewayb

CMD ["/gatewayb"]


# syntax=docker/dockerfile:1
FROM golang:1.22
WORKDIR gateway

# restore dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN apt update 

RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway

CMD ["/gateway"]


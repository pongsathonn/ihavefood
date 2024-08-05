# syntax=docker/dockerfile:1

FROM golang:1.22
WORKDIR coupon

# restore dependencies
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-coupon

CMD ["/docker-coupon"]


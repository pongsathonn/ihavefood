# syntax=docker/dockerfile:1

FROM golang:1.22
WORKDIR coupon

# restore dependencies
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /docker-coupon

CMD ["/docker-coupon"]


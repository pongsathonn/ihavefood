# syntax=docker/dockerfile:1
FROM golang:1.22
WORKDIR auth

# restore dependencies
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-auth

CMD ["/docker-auth"]


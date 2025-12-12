# syntax=docker/dockerfile:1
FROM  golang:1.24 AS builder
WORKDIR /src/auth
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags=all="-N -l" -o /auth

FROM alpine:latest
COPY --from=builder /auth /auth
COPY --from=builder /src/auth/db/migrations /db/migrations
CMD ["/auth"]



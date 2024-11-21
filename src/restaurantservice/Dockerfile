# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR restaurant

# restore dependencies
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /restaurant

# Stage 2: Runtime
FROM alpine:latest
COPY --from=builder /restaurant /restaurant

# for development stage(debugging)
COPY --from=builder /go/bin/dlv /go/bin/dlv
ENV PATH=$PATH:/go/bin

CMD ["/restaurant"]


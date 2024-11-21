# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder

# Set working directory
WORKDIR profile

# Copy and restore dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the application code
COPY . .
RUN go mod tidy

# Install Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Build the application with debugging symbols
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /profile

# Stage 2: Runtime
FROM alpine:latest
COPY --from=builder /profile /profile

# for development stage(debugging)
COPY --from=builder /go/bin/dlv /go/bin/dlv
ENV PATH=$PATH:/go/bin

CMD ["/profile"]

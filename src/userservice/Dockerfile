# syntax=docker/dockerfile:1
FROM golang:1.22

# Set working directory
WORKDIR user

# Install Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Copy and restore dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the application code
COPY . .
RUN go mod tidy

# Build the application with debugging symbols
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /docker-user

CMD ["/docker-user"]

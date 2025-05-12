# syntax=docker/dockerfile:1
FROM golang:1.22 AS builder
WORKDIR /src
<<<<<<< HEAD
=======

# restore dependencies
>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy
<<<<<<< HEAD
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go build -o /gateway

FROM alpine:latest
=======

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go build -o /gateway

# Stage 2: Runtime
FROM alpine:latest

>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
RUN apk add --no-cache libc6-compat
COPY --from=builder /gateway /gateway
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv

<<<<<<< HEAD
=======

>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
CMD ["/gateway"]



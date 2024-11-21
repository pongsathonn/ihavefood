# syntax=docker/dockerfile:1

FROM golang:1.22 as builder
WORKDIR /delivery

# restore dependencies
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN go mod tidy

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /delivery

# Stage 2: Runtime
FROM alpine:latest
COPY --from=builder /delivery /delivery

# for development stage(debugging)
COPY --from=builder /go/bin/dlv /go/bin/dlv
ENV PATH=$PATH:/go/bin

CMD ["/delivery"]

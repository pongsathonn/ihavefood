# syntax=docker/dockerfile:1

FROM rust:1.81 AS builder
WORKDIR /src/delivery
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
COPY . .
RUN cargo install --path .

FROM debian:bullseye-slim
COPY --from=builder /usr/local/cargo/bin/delivery /usr/local/bin/delivery

CMD ["/delivery"]

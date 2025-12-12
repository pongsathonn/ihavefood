FROM rust:1.87 AS build
WORKDIR /app
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    protobuf-compiler \
    libprotobuf-dev \
    libssl-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean

COPY Cargo.toml Cargo.lock ./
COPY genproto ./genproto
COPY . .
RUN cargo fetch

RUN cargo build --release && \
    cp target/release/deliveryservice /bin/deliveryservice

FROM debian:bookworm
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    libssl3 && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean
COPY --from=build /bin/deliveryservice /bin/deliveryservice
RUN useradd -m -u 1000 appuser && \
    chown appuser:appuser /bin/deliveryservice
USER appuser

CMD ["/bin/deliveryservice"]

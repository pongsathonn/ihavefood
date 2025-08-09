FROM rust:1.87-slim AS build
WORKDIR /app
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    protobuf-compiler \
    libprotobuf-dev \
    libssl-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean

RUN --mount=type=bind,source=src,target=src \
    --mount=type=bind,source=Cargo.toml,target=Cargo.toml \
    --mount=type=bind,source=Cargo.lock,target=Cargo.lock \
    --mount=type=bind,source=migrations,target=migrations \
    --mount=type=bind,source=genproto,target=genproto \
    --mount=type=bind,source=queries,target=queries \
    --mount=type=bind,source=.sqlx,target=.sqlx \
    --mount=type=cache,target=/app/target/ \
    --mount=type=cache,target=/usr/local/cargo/registry/ \
    cargo build --release && \
    cp target/release/deliveryservice /bin/deliveryservice

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get clean

COPY --from=build /bin/deliveryservice /bin/

CMD ["/bin/deliveryservice"]

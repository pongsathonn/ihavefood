mod broker;
mod database;
mod delivery;
mod grpc_impl;
mod models;

use anyhow::Result;

use broker::RabbitMQ;
use database::Db;
use delivery::MyDelivery;
use ihavefood::{
    customer_service_client::CustomerServiceClient, delivery_service_server::DeliveryServiceServer,
    merchant_service_client::MerchantServiceClient,
};
use lapin::{Connection, ConnectionProperties};
use log::info;
use sqlx::{
    sqlite::{SqliteConnectOptions, SqlitePool},
    Pool, Sqlite,
};
use std::net::{IpAddr, Ipv4Addr, SocketAddr};
use std::str::FromStr;
use std::sync::Arc;
use tokio::sync::Semaphore;
use tonic::transport::Server;

pub mod ihavefood {
    // tonic::include_proto!("ihavefood");
    include!("../genproto/ihavefood.rs");
}

async fn init_amqp_conn() -> Connection {
    let conn = Connection::connect(
        format!(
            "amqp://{}:{}@{}",
            dotenv::var("DELIVERY_AMQP_USER").expect("DELIVERY_AMQP_USER must be set"),
            dotenv::var("DELIVERY_AMQP_PASS").expect("DELIVERY_AMQP_PASS must be set"),
            dotenv::var("DELIVERY_AMQP_HOST").expect("DELIVERY_AMQP_HOST must be set"),
        )
        .as_str(),
        ConnectionProperties::default(),
    )
    .await
    .expect("Failed to connect AMQP");

    info!("AMQP connection established successfully");
    conn
}

async fn init_sqlite_pool() -> Pool<Sqlite> {
    let url = dotenv::var("DATABASE_URL").expect("DATABASE_URL must be set");

    let opts = SqliteConnectOptions::from_str(url.as_str())
        .expect("Failed to parse DATABASE_URL")
        .create_if_missing(true);

    let pool = SqlitePool::connect_with(opts)
        .await
        .expect("Failed to create Sqlite pool");

    sqlx::migrate!()
        .run(&pool)
        .await
        .expect("Failed to run sqlx migration");

    info!("SQLite connection pool initialized");
    pool
}

#[tokio::main]
async fn main() -> Result<()> {
    dotenv::dotenv().ok();
    std::env::set_var("RUST_LOG", "debug,lapin=warn");

    env_logger::builder()
        .format_file(true)
        .format_line_number(true)
        .format_target(false)
        .init();

    let customer_client = CustomerServiceClient::connect("customer:3333").await?;
    let merchant_client = MerchantServiceClient::connect("merchant:5555").await?;

    let app = MyDelivery {
        db: Arc::new(Db::new(init_sqlite_pool().await)),
        broker: Arc::new(RabbitMQ::new(init_amqp_conn().await)),
        task_limiter: Arc::new(Semaphore::new(100)),
        customercl: customer_client,
        merchantcl: merchant_client,
    };

    let socket = SocketAddr::new(
        IpAddr::V4(Ipv4Addr::UNSPECIFIED),
        dotenv::var("DELIVERY_SERVER_PORT")?.parse()?,
    );

    let server = Server::builder()
        .add_service(DeliveryServiceServer::new(app.clone()))
        .serve(socket);

    tokio::spawn(async move {
        app.clone().start_services().await.unwrap();
    });

    info!(
        "Server initialized and listening on {}",
        dotenv::var("DELIVERY_SERVER_PORT")?
    );

    server.await?;
    Ok(())
}

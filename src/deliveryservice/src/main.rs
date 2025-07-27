mod broker;
mod database;
mod delivery;
mod grpc_impl;
mod models;

use anyhow::Result;
use broker::RabbitMQ;
use database::Db;
use delivery::MyDelivery;
use ihavefood::delivery_service_server::DeliveryServiceServer;
use lapin::{Connection, ConnectionProperties};
use log::info;
use sqlx::{sqlite::SqlitePoolOptions, Pool, Sqlite};
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

// TODO: connect to sqlite engine
async fn init_sqlite_pool() -> Pool<Sqlite> {
    dotenv::dotenv().ok();

    let pool = SqlitePoolOptions::new()
        .max_connections(15)
        .connect(&dotenv::var("DATABASE_URL").expect("DELIVERY must be set"))
        .await
        .expect("Failed to create Sqlite connection pool");

    info!("SQLite connection pool initialized");
    pool
}

#[tokio::main]
async fn main() -> Result<()> {
    std::env::set_var("RUST_LOG", "debug,lapin=warn");

    env_logger::builder()
        .format_file(true)
        .format_line_number(true)
        .format_target(false)
        .init();

    let app = MyDelivery {
        db: Arc::new(Db::new(init_sqlite_pool().await)),
        broker: Arc::new(RabbitMQ::new(init_amqp_conn().await)),
        task_limiter: Arc::new(Semaphore::new(100)),
    };

    let server = Server::builder()
        .add_service(DeliveryServiceServer::new(app.clone()))
        .serve(dotenv::var("DELIVERY_URI")?.parse()?);

    tokio::spawn(async move {
        app.clone().start_services().await.unwrap();
    });

    info!(
        "Server initialized and listening on {}",
        dotenv::var("DELIVERY_URI")?
    );

    server.await?;
    Ok(())
}

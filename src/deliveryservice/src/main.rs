mod broker;
mod database;
mod delivery;
mod grpc_impl;
mod models;

use anyhow::Result;
use broker::RabbitMQ;
use database::Db;
use delivery::MyDelivery;
use env_logger::Env;
use ihavefood::delivery_service_server::DeliveryServiceServer;
use lapin::{Connection, ConnectionProperties};
use sqlx::{sqlite::SqlitePoolOptions, Pool, Sqlite};
use std::sync::Arc;
use tokio::sync::Semaphore;
use tonic::transport::Server;

pub mod ihavefood {
    // tonic::include_proto!("ihavefood");
    include!("../genproto/ihavefood.rs");
}

async fn init_amqp_conn() -> Connection {
    // temp =====================================
    std::env::set_var("DELIVERY_AMQP_USER", "guest");
    std::env::set_var("DELIVERY_AMQP_PASS", "guest");
    std::env::set_var("DELIVERY_AMQP_HOST", "localhost");
    std::env::set_var("DELIVERY_AMQP_PORT", "5672");
    //===============================================

    let conn = Connection::connect(
        format!(
            "amqp://{}:{}@{}:{}",
            dotenv::var("DELIVERY_AMQP_USER").expect("DELIVERY_AMQP_USER not set"),
            dotenv::var("DELIVERY_AMQP_PASS").expect("DELIVERY_AMQP_PASS not set"),
            dotenv::var("DELIVERY_AMQP_HOST").expect("DELIVERY_AMQP_HOST not set"),
            dotenv::var("DELIVERY_AMQP_PORT").expect("DELIVERY_AMQP_PORT not set"),
        )
        .as_str(),
        ConnectionProperties::default(),
    )
    .await
    .expect("Failed to connect AMQP");
    log::info!("AMQP connection established successfully");
    conn
}

// TODO: connect to sqlite engine
async fn init_sqlite_pool() -> Pool<Sqlite> {
    let pool = SqlitePoolOptions::new()
        .max_connections(15)
        .connect("sqlite::memory:")
        .await
        .expect("Failed to create SQLite connection pool");
    log::info!("SQLite connection pool initialized");
    pool
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::Builder::from_env(Env::default().default_filter_or("warn")).init();

    let server = Server::builder()
        .add_service(DeliveryServiceServer::new(MyDelivery {
            db: Arc::new(Db::new(init_sqlite_pool().await)),
            broker: Arc::new(RabbitMQ::new(init_amqp_conn().await)),
            task_limiter: Arc::new(Semaphore::new(100)),
        }))
        .serve(dotenv::var("DELIVERY_URI")?.parse()?);

    log::info!(
        "Server initialized and listening on {}",
        dotenv::var("DELIVERY_URI")?
    );

    server.await?;
    Ok(())
}

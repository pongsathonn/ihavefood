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
use sqlx::sqlite::SqlitePoolOptions;
use std::sync::Arc;
use tokio::sync::Semaphore;
use tonic::transport::Server;

pub mod ihavefood {
    // tonic::include_proto!("ihavefood");
    include!("../genproto/ihavefood.rs");
}

async fn init_amqp_conn() -> Connection {
    Connection::connect(
        format!(
            "amqp://{}:{}@{}:{}",
            dotenv::var("DELIVERY_AMQP_USER").unwrap(),
            dotenv::var("DELIVERY_AMQP_PASS").unwrap(),
            dotenv::var("DELIVERY_AMQP_HOST").unwrap(),
            dotenv::var("DELIVERY_AMQP_PORT").unwrap(),
        )
        .as_str(),
        ConnectionProperties::default(),
    )
    .await
    .unwrap()
}

#[tokio::main]
async fn main() -> Result<()> {
    // TODO: connect to sqlite engine
    let pool = SqlitePoolOptions::new()
        .max_connections(15)
        .connect("sqlite::memory:")
        .await?;

    Server::builder()
        .add_service(DeliveryServiceServer::new(MyDelivery {
            db: Arc::new(Db::new(pool)),
            broker: Arc::new(RabbitMQ::new(init_amqp_conn().await)),
            task_limiter: Arc::new(Semaphore::new(100)),
        }))
        .serve(dotenv::var("DELIVERY_URI").unwrap().parse()?)
        .await?;
    Ok(())
}

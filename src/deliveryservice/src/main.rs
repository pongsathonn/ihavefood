mod database_impl;
mod delivery_impl;
mod event_impl;
mod models;

use anyhow::Result;

use database_impl::Db;
use delivery_impl::MyDelivery;
use event_impl::*;
use ihavefood::{
    customer_service_client::CustomerServiceClient, delivery_service_server::DeliveryServiceServer,
    merchant_service_client::MerchantServiceClient,
};
use lapin::{Connection, ConnectionProperties};
use log::info;
use sqlx::{
    sqlite::{SqliteConnectOptions, SqlitePool},
    Sqlite,
};
use std::net::{IpAddr, Ipv4Addr, SocketAddr};
use std::str::FromStr;
use std::sync::Arc;
// use tokio::sync::Semaphore;
use tokio::time::Duration;
use tonic::transport::{Channel, Server};

pub mod ihavefood {
    // tonic::include_proto!("ihavefood");
    include!("../genproto/ihavefood.rs");
}

fn init_redis_pool() -> redis::Client {
    let client = redis::Client::open("redis://redisx:6379/").expect("Invalid Redis URL");
    let _ = client.get_connection();
    client
}

async fn init_amqp_conn() -> Connection {
    let conn = Connection::connect(
        format!(
            "amqp://{}:{}@{}",
            dotenv::var("RBMQ_DELIVERY_USER").expect("RBMQ_DELIVERY_USER must be set"),
            dotenv::var("RBMQ_DELIVERY_PASS").expect("RBMQ_DELIVERY_PASS must be set"),
            dotenv::var("AMQP_SERVER_URI").expect("AMQP_SERVER_URI must be set"),
        )
        .as_str(),
        ConnectionProperties::default(),
    )
    .await
    .expect("Failed to connect AMQP server");

    info!("AMQP connection established successfully");
    conn
}

async fn init_sqlite_pool() -> sqlx::Pool<Sqlite> {
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

async fn init_customer_client() -> Result<CustomerServiceClient<Channel>> {
    for _ in 0..5 {
        match CustomerServiceClient::connect(format!("http://{}", dotenv::var("CUSTOMER_URI")?))
            .await
        {
            Ok(conn) => return Ok(conn),
            Err(_) => {
                tokio::time::sleep(Duration::from_secs(5)).await;
                continue;
            }
        };
    }

    panic!("could not established customer client")
}

async fn init_merchant_client() -> Result<MerchantServiceClient<Channel>> {
    for _ in 0..5 {
        match MerchantServiceClient::connect(format!("http://{}", dotenv::var("MERCHANT_URI")?))
            .await
        {
            Ok(conn) => return Ok(conn),
            Err(_) => {
                tokio::time::sleep(Duration::from_secs(5)).await;
                continue;
            }
        };
    }

    panic!("could not established merchant client")
}

#[tokio::main]
async fn main() -> Result<()> {
    dotenv::dotenv().ok();

    std::env::set_var("RUST_LOG", "info,lapin=warn");

    env_logger::builder()
        .format_file(true)
        .format_line_number(true)
        .format_target(false)
        .init();

    let event_bus = Arc::new(EventBus::new(init_amqp_conn().await));
    let db = Arc::new(Db::new(init_sqlite_pool().await));
    let redis_cl = init_redis_pool();

    let my_delivery = MyDelivery {
        db: db.clone(),
        event_bus: event_bus.clone(),
        redis_cl: redis_cl.clone(),
        customercl: init_customer_client().await?,
        merchantcl: init_merchant_client().await?,
    };

    let socket = SocketAddr::new(
        IpAddr::V4(Ipv4Addr::UNSPECIFIED),
        dotenv::var("DELIVERY_SERVER_PORT")?.parse()?,
    );

    let server = Server::builder()
        .add_service(DeliveryServiceServer::new(my_delivery.clone()))
        .serve(socket);

    let event_bus_cloned = Arc::clone(&event_bus);
    let db_cloned = Arc::clone(&db);

    tokio::spawn(async {
        EventDispatcher::new(event_bus_cloned, db_cloned, redis_cl)
            .add_event(EventHandler {
                queue: String::from(""),
                key: String::from("order.placed.event"),
            })
            .add_event(EventHandler {
                queue: String::from(""),
                key: String::from("sync.rider.created"),
            })
            .run()
            .await
    });

    info!(
        "Server initialized and listening on {}",
        dotenv::var("DELIVERY_SERVER_PORT")?
    );

    server.await?;
    Ok(())
}

mod delivery_impl;
mod event_impl;
mod models;
mod mongo_impl;

use anyhow::Result;
use delivery_impl::MyDelivery;
use event_impl::*;
use ihavefood::{
    customer_service_client::CustomerServiceClient, delivery_service_server::DeliveryServiceServer,
    merchant_service_client::MerchantServiceClient,
};
use lapin::{Connection, ConnectionProperties};
use log::info;
use mongo_impl::Db;
use mongodb::{Client, Database};
use std::net::{IpAddr, Ipv4Addr, SocketAddr};
use std::sync::Arc;
// use tokio::sync::Semaphore;
use tokio::time::Duration;
use tonic::transport::{Channel, Server};

pub mod ihavefood {
    // tonic::include_proto!("ihavefood");
    include!("../genproto/ihavefood.rs");
}

fn init_redis_pool() -> Result<redis::Client> {
    let url = dotenv::var("REDIS_URL")?;
    let client = redis::Client::open(url)?;
    let mut conn = client.get_connection()?;
    redis::cmd("PING").query::<String>(&mut conn)?;
    info!("Redis connection successful!");
    Ok(client)
}

async fn init_amqp_conn() -> Result<Connection> {
    let conn = Connection::connect(
        format!(
            "amqp://{}:{}@{}/{}",
            dotenv::var("RBMQ_USER").expect("RBMQ_USER must be set"),
            dotenv::var("RBMQ_PASS").expect("RBMQ_PASS must be set"),
            dotenv::var("RBMQ_HOST").expect("RBMQ_HOST must be set"),
            dotenv::var("RBMQ_USER").expect("RBMQ_USER must be set"),
        )
        .as_str(),
        ConnectionProperties::default(),
    )
    .await?;

    info!("AMQP connection established successfully");
    Ok(conn)
}

async fn init_mongo_db() -> Result<Database> {
    let uri = format!(
        "mongodb+srv://{}:{}@{}/?appName={}",
        dotenv::var("MONGO_USER").expect("MONGO_USER must be set"),
        dotenv::var("MONGO_PASS").expect("MONGO_PASS must be set"),
        dotenv::var("MONGO_HOST").expect("MONGO_HOST must be set"),
        dotenv::var("MONGO_CLUSTER").expect("MONGO_CLUSTER must be set"),
    );
    let client = Client::with_uri_str(uri).await?;
    let database = client.database("deliverydb");
    return Ok(database);
}

async fn init_customer_client() -> Result<CustomerServiceClient<Channel>> {
    for _ in 0..5 {
        match CustomerServiceClient::connect(dotenv::var("CUSTOMER_URI")?).await {
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
        match MerchantServiceClient::connect(dotenv::var("MERCHANT_URI")?).await {
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

    let mongo_db = init_mongo_db().await?;

    let event_bus = Arc::new(EventBus::new(init_amqp_conn().await?));
    let db = Arc::new(Db::new(
        mongo_db.collection("deliveries"),
        mongo_db.collection("riders"),
    ));
    let redis_cl = init_redis_pool()?;

    let my_delivery = MyDelivery {
        event_bus: event_bus.clone(),
        redis_cl: redis_cl.clone(),
        customercl: init_customer_client().await?,
        merchantcl: init_merchant_client().await?,
    };

    let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
    health_reporter
        .set_service_status("", tonic_health::ServingStatus::Serving)
        .await;

    let socket = SocketAddr::new(
        IpAddr::V4(Ipv4Addr::UNSPECIFIED),
        dotenv::var("PORT")?.parse()?,
    );

    let server = Server::builder()
        .add_service(DeliveryServiceServer::new(my_delivery.clone()))
        .add_service(health_service)
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
        dotenv::var("PORT")?
    );

    server.await?;
    Ok(())
}

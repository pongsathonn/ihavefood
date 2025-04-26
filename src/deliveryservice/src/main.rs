mod database;
mod models;
mod rabbitmq;

use anyhow::{ensure, Result};
use database::Db;
use futures::StreamExt;
use ihavefood::delivery_service_server::{DeliveryService, DeliveryServiceServer};
use ihavefood::*;
use lapin::{options::BasicAckOptions, Connection, ConnectionProperties};
use log::error;
use models::*;
use prost::Message;
use rabbitmq::RabbitMQ;
use sqlx::sqlite::SqlitePoolOptions;
use tokio::sync::mpsc;
use tokio::time::{sleep, Duration};
use tokio_stream::wrappers::ReceiverStream;
use tonic::{transport::Server, Code, Request, Response, Status};

pub mod ihavefood {
    tonic::include_proto!("ihavefood");
}

#[derive(Debug)]
pub struct MyDelivery {
    db: Db,
    rabbitmq: RabbitMQ,
}

impl MyDelivery {
    async fn event_listening(&self) {
        let mut consumer = self
            .rabbitmq
            .subscribe("rider.assign.queue", "order.placed.event")
            .await;

        // Read as:
        // while consumer.next`Option<T>` has Some(delivery) then does {}
        // if delivery`Result<T,E>` is Ok(v) then does {}
        while let Some(delivery) = consumer.next().await {
            if let Ok(delivery) = delivery {
                if let Some(content_type) = delivery.properties.content_type() {
                    if content_type.as_str() != "application/json" {
                        // TODO: error
                    }

                    let _ = PlaceOrder::decode(delivery.data.as_ref()).unwrap();

                delivery.ack(BasicAckOptions::default()).await.unwrap();
            }
        }
    }
}

#[tonic::async_trait]
impl DeliveryService for MyDelivery {
    type GetOrderTrackingStream = ReceiverStream<Result<GetOrderTrackingResponse, Status>>;

    // UNTEST !!!!!!!!
    async fn get_order_tracking(
        &self,
        request: Request<GetOrderTrackingRequest>,
    ) -> Result<Response<Self::GetOrderTrackingStream>, Status> {
        let _ = request;
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            for _ in 1..6 {
                sleep(Duration::from_secs(5)).await;

                // TODO: tracking rider location from GoogleAPI or database
                tx.send(Ok(GetOrderTrackingResponse {
                    ..Default::default()
                }))
                .await
                .unwrap();
            }
        });
        Ok(Response::new(ReceiverStream::new(rx)))
    }

    async fn get_delivery_fee(
        &self,
        request: Request<GetDeliveryFeeRequest>,
    ) -> Result<Response<GetDeliveryFeeResponse>, Status> {
        let restaurant_point = Point {
            latitude: request.get_ref().restaurant_lat,
            longitude: request.get_ref().restaurant_long,
        };

        let user_point = Point {
            latitude: request.get_ref().user_lat,
            longitude: request.get_ref().user_long,
        };

        let delivery_fee = calc_delivery_fee(&user_point, &restaurant_point).map_err(|err| {
            error!("Error: {err}");
            Status::new(Code::Internal, "failed to calculate delivery fee")
        })?;

        Ok(Response::new(GetDeliveryFeeResponse { delivery_fee }))
    }

    async fn confirm_rider_accept(
        &self,
        request: Request<ConfirmRiderAcceptRequest>,
    ) -> Result<Response<PickupInfo>, Status> {
        let delivery = self
            .db
            .get_delivery(request.get_ref().order_id.as_str())
            .await
            .unwrap();

        match delivery.status {
            DbDeliveryStatus::Unaccept => (),
            DbDeliveryStatus::Delivered => {
                return Err(Status::invalid_argument("rider already accepted"))
            }
            DbDeliveryStatus::Accepted => {
                return Err(Status::invalid_argument("order already delivered"))
            }
        }

        // TODO: push notify rider has accepted the order

        self.db
            .update_delivery_rider(
                request.get_ref().order_id.as_str(),
                request.get_ref().rider_id.as_str(),
            )
            .await
            .unwrap();

        self.db
            .update_delivery_status(
                request.get_ref().order_id.as_str(),
                DbDeliveryStatus::Accepted,
            )
            .await
            .unwrap();

        Ok(Response::new(PickupInfo {
            pickup_code: delivery.pickup_code,
            pickup_location: Some(Point {
                latitude: delivery.pickup_location.latitude,
                longitude: delivery.pickup_location.longitude,
            }),
            drop_off_location: Some(Point {
                latitude: delivery.drop_off_location.latitude,
                longitude: delivery.drop_off_location.longitude,
            }),
        }))
    }

    async fn confirm_order_deliver(
        &self,
        request: Request<ConfirmOrderDeliverRequest>,
    ) -> Result<Response<()>, Status> {
        self.db
            .update_delivery_status(
                request.into_inner().order_id.as_str(),
                DbDeliveryStatus::Delivered,
            )
            .await
            .unwrap();
        Ok(Response::new(()))
    }
}

fn calc_delivery_fee(user_p: &Point, restau_p: &Point) -> Result<i32> {
    //distance(kilometers)
    let distance = haversine_distance(user_p, restau_p);

    ensure!(
        (0.0..=25.0).contains(&distance),
        "distance must be between 0km and 25km"
    );

    let delivery_fee: i32 = match distance {
        d if d <= 5.0 => 0,
        d if d <= 10.0 => 50,
        _ => 100,
    };

    Ok(delivery_fee)
}

// haversineDistance calculates the distance between two geographic points in kilometers.
fn haversine_distance(p1: &Point, p2: &Point) -> f64 {
    // Earth's radius in kilometers.
    const EARTH_RADIUS: f64 = 6371.0;

    // Convert latitude and longitude from degrees to radians.
    let lat1 = p1.latitude.to_radians();
    let lon1 = p1.longitude.to_radians();
    let lat2 = p2.latitude.to_radians();
    let lon2 = p2.longitude.to_radians();

    // Differences in coordinates
    let dlat = lat2 - lat1;
    let dlon = lon2 - lon1;

    // Haversine formula
    let a = (dlat / 2.0).sin().powi(2) + lat1.cos() * lat2.cos() * (dlon / 2.0).sin().powi(2);
    let c = 2.0 * a.sqrt().asin();

    // Calculate the distance
    EARTH_RADIUS * c
}

async fn init_amqp_conn() -> Connection {
    let addr = "amqp://127.0.0.1:5672";
    Connection::connect(addr, ConnectionProperties::default())
        .await
        .unwrap()
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::1]:5555".parse()?;

    // TODO : connect to sqlite engine
    let pool = SqlitePoolOptions::new()
        .max_connections(15)
        .connect("sqlite::memory:")
        .await?;

    Server::builder()
        .add_service(DeliveryServiceServer::new(MyDelivery {
            db: Db::new(pool),
            rabbitmq: RabbitMQ::new(init_amqp_conn().await),
        }))
        .serve(addr)
        .await?;
    Ok(())
}

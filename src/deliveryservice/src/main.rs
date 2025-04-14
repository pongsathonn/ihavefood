mod database;
mod models;

use anyhow::{ensure, Result};
use ihavefood::delivery_service_server::{DeliveryService, DeliveryServiceServer};
use ihavefood::{
    ConfirmOrderDeliverRequest, ConfirmRiderAcceptRequest, GetDeliveryFeeRequest,
    GetDeliveryFeeResponse, GetOrderTrackingRequest, GetOrderTrackingResponse, PickupInfo, Point,
};
use log::error;

use tokio::sync::mpsc;
use tokio::time::{sleep, Duration};
use tokio_stream::wrappers::ReceiverStream;
use tonic::{transport::Server, Code, Request, Response, Status};

use database::Db;
use models::*;
use sqlx::sqlite::SqlitePoolOptions;

// _____ ___  ____   ___       __     __    _ _     _       _
//|_   _/ _ \|  _ \ / _ \   _  \ \   / /_ _| (_) __| | __ _| |_ ___
//  | || | | | | | | | | | (_)  \ \ / / _` | | |/ _` |/ _` | __/ _ \
//  | || |_| | |_| | |_| |  _    \ V / (_| | | | (_| | (_| | ||  __/
//  |_| \___/|____/ \___/  (_)    \_/ \__,_|_|_|\__,_|\__,_|\__\___|

pub mod ihavefood {
    tonic::include_proto!("ihavefood");
}

#[derive(Debug)]
pub struct MyDelivery {
    db: Db,
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

        if delivery.status != DbDeliveryStatus::UNACCEPT {
            return Err(Status::new(Code::InvalidArgument, "invalid status"));
        } else if delivery.status == DbDeliveryStatus::DELIVERED {
            return Err(Status::new(
                Code::InvalidArgument,
                "order already delivered",
            ));
        };

        self.db
            .update_delivery_rider(&DbUpdateDeliveryRider {
                order_id: request.get_ref().order_id.clone(),
                rider_id: request.get_ref().rider_id.clone(),
                accept_time: chrono::Utc::now(),
                status: DbDeliveryStatus::ACCEPTED,
            })
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
        // just update order status

        _ = request;
        todo!()
    }
}

fn calc_delivery_fee(user_p: &Point, restau_p: &Point) -> Result<i32> {
    //distance(kilometers)
    let distance = haversine_distance(user_p, restau_p);

    ensure!(
        distance >= 0.0 && distance <= 25.0,
        "distance must be between 0km and 25km"
    );

    let delivery_fee: i32;
    match distance {
        d if d <= 5.0 => delivery_fee = 0,
        d if d <= 10.0 => delivery_fee = 50,
        _ => delivery_fee = 100,
    }
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

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "[::1]:5555".parse()?;

    // TODO : connect to sqlite engine
    let pool = SqlitePoolOptions::new()
        .max_connections(15)
        .connect("sqlite::memory:")
        .await?;

    Server::builder()
        .add_service(DeliveryServiceServer::new(MyDelivery { db: Db::new(pool) }))
        .serve(addr)
        .await?;
    Ok(())
}

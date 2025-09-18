use crate::ihavefood::delivery_service_server::DeliveryService;
use crate::ihavefood::*;
use crate::models::*;
use crate::Db;
use crate::EventBus;

use self::{
    customer_service_client::CustomerServiceClient, merchant_service_client::MerchantServiceClient,
};
use anyhow::{anyhow, bail, ensure, Context, Result};
use chrono::Utc;
use log::error;
use prost::Message;
use rand::prelude::*;
// use redis::Commands;
use redis::AsyncCommands;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::mpsc;
use tokio::time::{sleep, Duration};
use tokio_stream::wrappers::ReceiverStream;
use tonic::transport::Channel;
use tonic::{Code, Request, Response, Status};

#[derive(Debug, Clone)]
pub struct MyDelivery {
    pub db: Arc<Db>,
    pub event_bus: Arc<EventBus>,
    pub redis_cl: redis::Client,

    pub customercl: CustomerServiceClient<Channel>,
    pub merchantcl: MerchantServiceClient<Channel>,
}

impl MyDelivery {
    pub fn prepare_order_delivery(order: &PlaceOrder) -> Result<(Vec<Rider>, PickupInfo)> {
        let riders = Self::calc_nearest_riders();
        let pickup_info = Self::generate_order_pickup(order)?;
        Ok((riders, pickup_info))
    }

    pub fn notify_riders(riders: Vec<Rider>, pickup_info: PickupInfo) -> Result<()> {
        // TODO: query neasted riders and notify
        for rider in riders.iter() {
            let rider_id = rider.rider_id.as_str();
            let code = pickup_info.pickup_code.as_str();
            log::info!("notified to rider={rider_id}, pickup_code={code}");
        }
        Ok(())
    }

    pub fn calc_delivery_fee(customer_p: &Point, restau_p: &Point) -> Result<i32> {
        //distance(kilometers)
        let distance = haversine_distance(customer_p, restau_p);

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

    // TODO: use actual calculation format
    fn calc_nearest_riders() -> Vec<Rider> {
        let mut riders: Vec<Rider> = Vec::new();
        (0..5).for_each(|_| riders.push(Rider::default()));
        riders
    }

    fn generate_order_pickup(order: &PlaceOrder) -> Result<PickupInfo> {
        let pickup_code = match (100..1000).collect::<Vec<i32>>().choose(&mut rand::rng()) {
            Some(code) => code.to_string(),
            None => {
                bail!("Failed to generate pickup code");
            }
        };

        let pickup_location = address_to_point(
            order
                .merchant_address
                .as_ref()
                .ok_or_else(|| anyhow!("Merchant address is empty"))?,
        );

        let drop_off_location = address_to_point(
            order
                .customer_address
                .as_ref()
                .ok_or_else(|| anyhow!("User address is empty"))?,
        );

        Ok(PickupInfo {
            pickup_code,
            pickup_location,
            drop_off_location,
        })
    }
}

#[tonic::async_trait]
impl DeliveryService for MyDelivery {
    type TrackingRiderStream = ReceiverStream<Result<TrackingRiderResponse, Status>>;

    // TODO: tracking order or 'Rider'
    //
    // might create another method for tracking domain logic like
    // - tracking rider picking up an order yet ?
    // - tracking is rider delivered yet ?
    //
    // cause this function just tracking rider location
    async fn tracking_rider(
        &self,
        request: Request<TrackingRiderRequest>,
    ) -> Result<Response<Self::TrackingRiderStream>, Status> {
        let _ = request;
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            for _ in 1..6 {
                sleep(Duration::from_secs(5)).await;

                // TODO: tracking rider location from GoogleAPI
                if tx
                    .send(Ok(TrackingRiderResponse {
                        ..Default::default()
                    }))
                    .await
                    .is_err()
                {
                    error!("receiver dropped");
                    return;
                }
            }
        });

        Ok(Response::new(ReceiverStream::new(rx)))
    }

    async fn get_delivery_fee(
        &self,
        request: Request<GetDeliveryFeeRequest>,
    ) -> Result<Response<GetDeliveryFeeResponse>, Status> {
        let customer = self
            .customercl
            .clone()
            .get_customer(GetCustomerRequest {
                customer_id: request.get_ref().customer_id.clone(),
            })
            .await?;

        let customer_addr = customer
            .get_ref()
            .addresses
            .iter()
            .find(|&addr| addr.address_id == request.get_ref().customer_address_id)
            .ok_or(Status::new(Code::Internal, "internal server error"))?;

        let merchant = self
            .merchantcl
            .clone()
            .get_merchant(GetMerchantRequest {
                merchant_id: request.get_ref().merchant_id.clone(),
            })
            .await?;

        let merchant_addr = merchant
            .get_ref()
            .address
            .as_ref()
            .ok_or(Status::new(Code::Internal, "internal server error"))?;

        let customer_point = fake_geocode(customer_addr);
        let merchant_point = fake_geocode(merchant_addr);
        let fee = Self::calc_delivery_fee(&customer_point, &merchant_point).map_err(|err| {
            error!("calculate delivery fee: {err}");
            Status::new(Code::Internal, "failed to calculate delivery fee")
        })?;

        Ok(Response::new(GetDeliveryFeeResponse { fee }))
    }

    async fn report_delivery_status(
        &self,
        request: Request<ReportDeliveryStatusRequest>,
    ) -> Result<Response<::prost_wkt_types::Empty>, Status> {
        let order_id = &request.get_ref().order_id;
        let rider_id = &request.get_ref().rider_id;
        let new_status = request.get_ref().status();

        if order_id.is_empty() {
            return Err(Status::invalid_argument("Order ID cannot be empty"));
        }
        if rider_id.is_empty() {
            return Err(Status::invalid_argument("Rider ID cannot be empty"));
        }
        if new_status.eq(&DeliveryStatus::RiderUnaccept) {
            return Err(Status::invalid_argument("Status should not be UNACCEPT"));
        }

        let mut redis = match self.redis_cl.get_multiplexed_async_connection().await {
            Ok(redis_conn) => redis_conn,
            Err(e) => {
                error!("Failed to estrablish Redis connection: {e}");
                return Err(Status::internal("server internal error"));
            }
        };

        // TEST: Set initial status
        // let _: () = redis
        //     .hset(
        //         order_id,
        //         "status",
        //         DeliveryStatus::RiderAccepted.as_str_name(),
        //     )
        //     .await
        //     .map_err(|e| {
        //         error!("Failed to set order status in Redis: {:?}", e);
        //         Status::internal("server internal error")
        //     })?;

        let current_status = redis
            .hget(order_id, "status")
            .await
            .map_err(|e| {
                error!("Failed to retrieve order status from Redis: {:?}", e);
                Status::internal("server internal error")
            })
            .and_then(|v: String| {
                DeliveryStatus::from_str_name(&v)
                    .ok_or_else(|| Status::internal("invalid status value"))
            })?;

        match new_status {
            DeliveryStatus::RiderUnaccept => {
                error!("status validation not implement");
                return Err(Status::internal("internal server error"));
            }
            DeliveryStatus::RiderAccepted => {
                if current_status == DeliveryStatus::RiderAccepted
                    || current_status == DeliveryStatus::RiderPickedUp
                    || current_status == DeliveryStatus::RiderDelivered
                {
                    return Err(Status::failed_precondition("order has already accepted"));
                }
            }
            DeliveryStatus::RiderPickedUp => {
                if current_status == DeliveryStatus::RiderDelivered
                    || current_status == DeliveryStatus::RiderPickedUp
                {
                    return Err(Status::failed_precondition("order has already picked up"));
                }
            }
            DeliveryStatus::RiderDelivered => {
                if current_status == DeliveryStatus::RiderDelivered {
                    return Err(Status::failed_precondition("order has already delivered"));
                }
            }
        };

        let _: () = redis
            .hset(order_id, "status", new_status.as_str_name())
            .await
            .map_err(|e| {
                error!("Failed to update delivery status : {e}");
                Status::internal("server internal error")
            })?;

        let event = RiderAssignedEvent {
            order_id: order_id.clone(),
            rider_id: rider_id.clone(),
            assign_time: Some(prost_wkt_types::Timestamp::from(Utc::now())),
        };

        self.event_bus
            .publish("rider.assigned.event", &event.encode_to_vec())
            .await
            .map_err(|e| {
                error!("Failed to publish rider assigned event: {:?}", e);
                Status::internal("server internal error")
            })?;

        Ok(Response::new(::prost_wkt_types::Empty {}))
    }
}

// convert Address to Point.
//
// TODO: implememnt Geocoding ( Google APIs )
fn address_to_point(address: &Address) -> Option<Point> {
    // [ Chaing Mai district ]
    let example: HashMap<&str, Point> = HashMap::from([
        (
            "Mueang",
            Point {
                latitude: 18.7883,
                longitude: 98.9853,
            },
        ),
        (
            "Hang Dong",
            Point {
                latitude: 18.6870,
                longitude: 98.8897,
            },
        ),
        (
            "San Sai",
            Point {
                latitude: 18.8578,
                longitude: 99.0631,
            },
        ),
        (
            "Mae Rim",
            Point {
                latitude: 18.8998,
                longitude: 98.9311,
            },
        ),
        (
            "Doi Saket",
            Point {
                latitude: 18.8482,
                longitude: 99.1403,
            },
        ),
    ]);

    example.get(address.district.as_str()).cloned()
}

// haversineDistance calculates the distance between two geographic points in kilometers.
fn haversine_distance(p1: &Point, p2: &Point) -> f64 {
    // Earth's radius in kilometers.
    const EARTH_RADIUS: f64 = 6371.0;

    let lat1 = p1.latitude.to_radians();
    let lon1 = p1.longitude.to_radians();
    let lat2 = p2.latitude.to_radians();
    let lon2 = p2.longitude.to_radians();

    let dlat = lat2 - lat1;
    let dlon = lon2 - lon1;

    // Haversine formula
    let a = (dlat / 2.0).sin().powi(2) + lat1.cos() * lat2.cos() * (dlon / 2.0).sin().powi(2);
    let c = 2.0 * a.sqrt().asin();

    EARTH_RADIUS * c
}

// fake_geocode from ChatGPT
pub fn fake_geocode(_addr: &Address) -> Point {
    let mut rng = rand::rng();

    // arbitrary “center” at 0,0 and offset within ~25 km (~0.225 lat, ~0.25 lng)
    let max_lat_offset = 0.225;
    let max_lng_offset = 0.25;

    Point {
        latitude: rng.random_range(-max_lat_offset..=max_lat_offset),
        longitude: rng.random_range(-max_lng_offset..=max_lng_offset),
    }
}

// #[cfg(test)]
// mod tests {
//     use super::*;
//     // use crate::models::*;
//     use serde_json;
//     use std::{fs::File, io::BufReader};
//
//     #[test]
//     fn test_calc_delivery_fee_success() {
//         let customer_point = Point {
//             latitude: 18.7883,
//             longitude: 98.9853,
//         };
//         let merchant_point = Point {
//             latitude: 18.6870,
//             longitude: 98.8897,
//         };
//
//         let delivery_fee = MyDelivery::calc_delivery_fee(&customer_point, &merchant_point).unwrap();
//         println!("{}", delivery_fee);
//         assert_eq!(delivery_fee, 50);
//     }
//
//     #[test]
//     fn test_calc_delivery_fee_failure() {
//         let customer_point = Point {
//             latitude: 18.7883,
//             longitude: 98.9853,
//         };
//         let merchant_point = Point {
//             latitude: 50.0000,
//             longitude: 50.0000,
//         };
//
//         let result = MyDelivery::calc_delivery_fee(&customer_point, &merchant_point);
//
//         assert!(result.is_err());
//         assert!(result
//             .unwrap_err()
//             .to_string()
//             .contains("distance must be between 0km and 25km"));
//     }
//
//     #[test]
//     fn test_prepare_order_delivery_success() {
//         // PlaceOrder needs serde deserialize (derive)
//         let place_orders: Vec<PlaceOrder> = serde_json::from_reader(BufReader::new(
//             File::open("/testdata/place_order_success.json").unwrap(),
//         ))
//         .unwrap();
//
//         println!("placeorder ja = {:?}", place_orders);
//
//         for place_order in place_orders {
//             let result = MyDelivery::prepare_order_delivery(&place_order);
//
//             assert!(result.is_ok());
//             let (riders, pickup_info) = result.unwrap();
//             assert_eq!(riders.len(), 5);
//             assert!(pickup_info.pickup_location.is_some());
//             assert!(pickup_info.drop_off_location.is_some());
//         }
//     }

// #[test]
// fn test_prepare_order_delivery_failure() {
//     let order = PlaceOrder {
//         order_id: "test_order_456".to_string(),
//         merchant_address: None, // Missing address
//         customer_address: Some(Address {
//             district: "Hang Dong".to_string(),
//             street: "User Street".to_string(),
//             building: "User Building".to_string(),
//         }),
//     };
//
//     let result = MyDelivery::prepare_order_delivery(&order);
//
//     assert!(result.is_err());
//     assert!(result
//         .unwrap_err()
//         .to_string()
//         .contains("Merchant address is empty"));
// }
//
// #[test]
// fn test_address_to_point_success() {
//     let address = Address {
//         district: "Mueang".to_string(),
//         street: "Test Street".to_string(),
//         building: "Test Building".to_string(),
//     };
//
//     let result = address_to_point(&address);
//
//     assert!(result.is_some());
//     let point = result.unwrap();
//     assert_eq!(point.latitude, 18.7883);
//     assert_eq!(point.longitude, 98.9853);
// }
//
// #[test]
// fn test_address_to_point_failure() {
//     let address = Address {
//         district: "Unknown District".to_string(), // Not in the HashMap
//         street: "Test Street".to_string(),
//         building: "Test Building".to_string(),
//     };
//
//     let result = address_to_point(&address);
//
//     assert!(result.is_none());
// }
//
// #[test]
// fn test_haversine_distance_success() {
//     let p1 = Point {
//         latitude: 18.7883,
//         longitude: 98.9853,
//     };
//     let p2 = Point {
//         latitude: 18.6870,
//         longitude: 98.8897,
//     };
//
//     let distance = haversine_distance(&p1, &p2);
//
//     assert!(distance > 0.0);
//     assert!(distance < 25.0); // Should be reasonable distance
//     assert!((distance - 12.0).abs() < 5.0); // Roughly 12km +/- 5km tolerance
// }
//
// #[test]
// fn test_haversine_distance_failure() {
//     let p1 = Point {
//         latitude: 0.0,
//         longitude: 0.0,
//     };
//     let p2 = Point {
//         latitude: 0.0,
//         longitude: 0.0,
//     };
//
//     let distance = haversine_distance(&p1, &p2);
//
//     // Same point should return 0 distance
//     assert_eq!(distance, 0.0);
// }
//
// #[test]
// fn test_calc_nearest_riders_success() {
//     let riders = MyDelivery::calc_nearest_riders();
//
//     assert_eq!(riders.len(), 5);
//     // All riders should be default instances
//     for rider in riders {
//         assert_eq!(rider, Rider::default());
//     }
// }
//
// #[test]
// fn test_calc_nearest_riders_failure() {
//     // This function always returns 5 riders, so we test the constraint
//     let riders = MyDelivery::calc_nearest_riders();
//
//     // Failure case: should not return empty or wrong count
//     assert_ne!(riders.len(), 0);
//     assert_ne!(riders.len(), 10); // Not the expected 5
// }
//
// #[test]
// fn test_notify_riders_success() {
//     let riders = vec![Rider::default(), Rider::default()];
//
//     // This function just logs, so we test it doesn't panic
//     MyDelivery::notify_riders(riders);
//
//     // If we reach here, it succeeded
//     assert!(true);
// }
//
// #[test]
// fn test_notify_riders_failure() {
//     let riders = vec![]; // Empty riders list
//
//     // Should still not panic with empty list
//     MyDelivery::notify_riders(riders);
//
//     // Test passes if no panic occurs
//     assert!(true);
// }
// }

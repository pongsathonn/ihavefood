use crate::broker::RabbitMQ;
use crate::database::Db;
use crate::ihavefood::*;
use crate::models::*;

use anyhow::{anyhow, bail, ensure, Result};
use futures::StreamExt;
use lapin::options::BasicAckOptions;
use log::error;
use prost::Message;
use rand::prelude::*;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::Semaphore;

#[derive(Debug, Clone)]
pub struct MyDelivery {
    pub db: Arc<Db>,
    pub broker: Arc<RabbitMQ>,
    pub task_limiter: Arc<Semaphore>,
}

impl MyDelivery {
    pub async fn start_services(&self) -> Result<()> {
        tokio::try_join!(self.delivery_assignment(), self.other_event_handler())?;
        Ok(())
    }

    async fn other_event_handler(&self) -> Result<()> {
        Ok(())
    }

    async fn delivery_assignment(&self) -> Result<()> {
        let mut consumer = match self
            .broker
            .subscribe("rider.assign.queue", "order.placed.event")
            .await
        {
            Ok(consumer) => consumer,
            Err(e) => {
                error!("Failed to subscribe to delivery queue: {}", e);
                return Ok(());
            }
        };

        while let Some(delivery) = consumer.next().await {
            let db = Arc::clone(&self.db);
            let task_limiter = Arc::clone(&self.task_limiter);

            tokio::spawn(async move {
                let _permit = match task_limiter.acquire().await {
                    Ok(permit) => permit,
                    Err(e) => {
                        error!("Failed to acquire task permit: {}", e);
                        return;
                    }
                };

                if let Ok(delivery) = delivery {
                    if let Err(err) = delivery.ack(BasicAckOptions::default()).await {
                        error!("Failed to ack:{}", err);
                        return;
                    }

                    // TODO: check delivery id duplicated

                    let place_order = match PlaceOrder::decode(delivery.data.as_ref()) {
                        Ok(p) => p,
                        Err(e) => {
                            error!("Failed to decode place order:{}", e);
                            return;
                        }
                    };

                    let (riders, pickup_info) = match Self::prepare_order_delivery(&place_order) {
                        Ok(v) => v,
                        Err(e) => {
                            error!("Failed to prepare order:{}", e);
                            return;
                        }
                    };

                    let pickup_location = match pickup_info.pickup_location {
                        Some(location) => location,
                        None => {
                            error!("Empty pickup information");
                            return;
                        }
                    };

                    let drop_off_location = match pickup_info.drop_off_location {
                        Some(location) => location,
                        None => {
                            error!("Empty pickup information");
                            return;
                        }
                    };

                    if let Err(e) = db
                        .create_delivery(&NewDelivery {
                            order_id: place_order.order_id,
                            pickup_code: pickup_info.pickup_code,
                            pickup_location: DbPoint {
                                latitude: pickup_location.latitude,
                                longitude: pickup_location.longitude,
                            },
                            drop_off_location: DbPoint {
                                latitude: drop_off_location.latitude,
                                longitude: drop_off_location.longitude,
                            },
                            create_time: chrono::Utc::now(),
                        })
                        .await
                    {
                        error!("Failed to create delivery record: {}", e);
                        return;
                    }

                    Self::notify_riders(riders);
                }
            });
        }

        Ok(())
    }

    async fn handle_rider_ack(&self, order_id: String, rider_id: String) -> Result<PickupInfo> {
        self.db
            .update_delivery_rider(&order_id, &rider_id)
            .await
            .unwrap();

        let delivery = self.db.get_delivery(&order_id).await.unwrap();

        self.broker
            .publish("rider.assigned.event", order_id.as_bytes())
            .await
            .unwrap();

        Ok(PickupInfo {
            pickup_code: delivery.pickup_code,
            pickup_location: Some(Point {
                latitude: delivery.pickup_location.latitude,
                longitude: delivery.pickup_location.longitude,
            }),
            drop_off_location: Some(Point {
                latitude: delivery.drop_off_location.latitude,
                longitude: delivery.drop_off_location.longitude,
            }),
        })
    }

    pub fn prepare_order_delivery(order: &PlaceOrder) -> Result<(Vec<Rider>, PickupInfo)> {
        let riders = Self::calc_nearest_riders();
        let pickup_info = Self::generate_order_pickup(order)?;
        Ok((riders, pickup_info))
    }

    // TODO: implement some notification
    fn notify_riders(_riders: Vec<Rider>) {
        log::info!("notify to riders");
    }

    pub fn calc_delivery_fee(user_p: &Point, restau_p: &Point) -> Result<i32> {
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
                .restaurant_address
                .as_ref()
                .ok_or_else(|| anyhow!("Restaurant address is empty"))?,
        );

        let drop_off_location = address_to_point(
            order
                .user_address
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

#[cfg(test)]
mod tests {
    use super::*;
    // use crate::models::*;
    use serde_json;
    use std::{fs::File, io::BufReader};

    #[test]
    fn test_calc_delivery_fee_success() {
        let user_point = Point {
            latitude: 18.7883,
            longitude: 98.9853,
        };
        let restaurant_point = Point {
            latitude: 18.6870,
            longitude: 98.8897,
        };

        let delivery_fee = MyDelivery::calc_delivery_fee(&user_point, &restaurant_point).unwrap();
        println!("{}", delivery_fee);
        assert_eq!(delivery_fee, 50);
    }

    #[test]
    fn test_calc_delivery_fee_failure() {
        let user_point = Point {
            latitude: 18.7883,
            longitude: 98.9853,
        };
        let restaurant_point = Point {
            latitude: 50.0000,
            longitude: 50.0000,
        };

        let result = MyDelivery::calc_delivery_fee(&user_point, &restaurant_point);

        assert!(result.is_err());
        assert!(result
            .unwrap_err()
            .to_string()
            .contains("distance must be between 0km and 25km"));
    }

    #[test]
    fn test_prepare_order_delivery_success() {
        // PlaceOrder needs serde deserialize (derive)
        let place_orders: Vec<PlaceOrder> = serde_json::from_reader(BufReader::new(
            File::open("/testdata/place_order_success.json").unwrap(),
        ))
        .unwrap();

        println!("placeorder ja = {:?}", place_orders);

        for place_order in place_orders {
            let result = MyDelivery::prepare_order_delivery(&place_order);

            assert!(result.is_ok());
            let (riders, pickup_info) = result.unwrap();
            assert_eq!(riders.len(), 5);
            assert!(pickup_info.pickup_location.is_some());
            assert!(pickup_info.drop_off_location.is_some());
        }
    }

    // #[test]
    // fn test_prepare_order_delivery_failure() {
    //     let order = PlaceOrder {
    //         order_id: "test_order_456".to_string(),
    //         restaurant_address: None, // Missing address
    //         user_address: Some(Address {
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
    //         .contains("Restaurant address is empty"));
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
}

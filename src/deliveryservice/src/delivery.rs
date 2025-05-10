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

#[derive(Debug)]
pub struct MyDelivery {
    pub db: Arc<Db>,
    pub broker: Arc<RabbitMQ>,
    pub task_limiter: Arc<Semaphore>,
}

impl MyDelivery {
    pub async fn start_services(&self) -> Result<()> {
        let other_task = self.other_event_handler();
        let delivery_task = self.delivery_assignment();
        tokio::try_join!(other_task, delivery_task,)?;
        Ok(())
    }

    async fn other_event_handler(&self) -> Result<()> {
        // unimplement
        Ok(())
    }

    async fn delivery_assignment(&self) -> Result<()> {
        let mut consumer = self
            .broker
            .subscribe("rider.assign.queue", "order.placed.event")
            .await;

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
                /////////////////////////////////////

                if let Ok(delivery) = delivery {
                    if let Err(err) = delivery.ack(BasicAckOptions::default()).await {
                        error!("Failed to ack:{}", err);
                        return;
                    }

                    // TODO: check delivery id duplicated

                    if let Some(content_type) = delivery.properties.content_type() {
                        if content_type.as_str() != "application/json" {
                            error!("Invalid content type{}:", content_type);
                            return;
                        }
                    } else {
                        error!("No content type");
                        return;
                    }

                    let place_order = match PlaceOrder::decode(delivery.data.as_ref()) {
                        Ok(p) => p,
                        Err(e) => {
                            error!("Failed to decode:{}", e);
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

    fn notify_riders(_riders: Vec<Rider>) {
        unimplemented!();
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

        let pickup_location = Self::address_to_point(
            order
                .restaurant_address
                .as_ref()
                .ok_or_else(|| anyhow!("Restaurant address is empty"))?,
        );

        let drop_off_location = Self::address_to_point(
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

        example.get(address.province.as_str()).cloned()
    }
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

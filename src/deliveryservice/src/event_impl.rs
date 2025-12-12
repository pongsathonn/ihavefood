use crate::ihavefood::*;
use crate::models::*;
use crate::Db;
use crate::MyDelivery;

use anyhow::{anyhow, Context, Result};
use bytes::Buf;
use chrono::Utc;
use futures::StreamExt;
use lapin::{options::*, types::FieldTable, BasicProperties, Connection, Consumer, ExchangeKind};
use log::error;
use prost::Message;
use redis::AsyncCommands;
use std::sync::Arc;

#[derive(Clone)]
pub struct EventHandler {
    pub queue: String,
    pub key: String,
}

#[derive(Debug)]
pub struct EventBus {
    conn: Connection,
}

#[derive(Clone)]
pub struct EventDispatcher {
    pub events: Vec<EventHandler>,
    pub event_bus: Arc<EventBus>,
    pub redis_cl: redis::Client,
    pub db: Arc<Db>,
}

impl EventBus {
    pub fn new(conn: Connection) -> Self {
        Self { conn }
    }

    pub async fn publish(&self, key: &str, payload: &[u8]) -> Result<()> {
        let ch = self.conn.create_channel().await?;

        ch.exchange_declare(
            "my_exchange",
            ExchangeKind::Direct,
            ExchangeDeclareOptions::default(),
            FieldTable::default(),
        )
        .await?;

        let confirm = ch
            .basic_publish(
                "my_exchange",
                key,
                BasicPublishOptions::default(),
                payload,
                BasicProperties::default(),
            )
            .await?
            .await?;

        if confirm.is_ack() {
            Ok(())
        } else {
            Err(anyhow::Error::msg("TODO: error"))
        }
    }

    // TODO: handle error properly
    pub async fn subscribe(&self, queue: &str, key: &str) -> Result<Consumer> {
        let ch = self.conn.create_channel().await?;

        ch.exchange_declare(
            "my_exchange",
            ExchangeKind::Direct,
            ExchangeDeclareOptions {
                passive: false,
                durable: true,
                auto_delete: false,
                internal: false,
                nowait: false,
            },
            FieldTable::default(),
        )
        .await?;

        ch.queue_declare(queue, QueueDeclareOptions::default(), FieldTable::default())
            .await?;

        ch.queue_bind(
            queue,
            "my_exchange",
            key,
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await?;

        ch.basic_consume(
            queue,
            "delivery_service",
            BasicConsumeOptions::default(),
            FieldTable::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!(e))
    }
}

impl EventDispatcher {
    pub fn new(event_bus: Arc<EventBus>, db: Arc<Db>, redis_cl: redis::Client) -> Self {
        Self {
            events: Vec::new(),
            event_bus,
            redis_cl,
            db,
        }
    }

    pub fn add_event(mut self, event: EventHandler) -> Self {
        self.events.push(event);
        self
    }

    pub async fn run(self) -> Result<()> {
        let self_loop = Arc::new(self);

        for event in self_loop.events.iter() {
            let key = event.key.clone();
            let self_cloned = Arc::clone(&self_loop);

            // TODO
            // - spawn new thread for each event
            // - use task_limiter (Semaphore)
            // let _permit = match task_limiter.acquire().await {
            // Ok(permit) => permit,
            // Err(e) => {
            //     error!("task limiter acquire: {}", e);
            //     return;
            // }

            let mut consumer = match self_loop
                .event_bus
                .subscribe(event.queue.as_str(), event.key.as_str())
                .await
            {
                Ok(consumer) => consumer,
                Err(err) => {
                    error!("Error: failed to subscribe: {err}");
                    continue;
                }
            };

            tokio::spawn(async move {
                while let Some(delivery) = consumer.next().await {
                    let data = delivery.unwrap().data;

                    match key.as_ref() {
                        "order.placed.event" => {
                            if let Err(e) = self_cloned.handle_order_placed(data.as_ref()).await {
                                error!("failed to handle order placed: {}", e);
                                return;
                            }
                        }
                        "sync.rider.created" => {
                            if let Err(e) = self_cloned.handle_rider_created(data.as_ref()).await {
                                error!("failed to handle rider created: {}", e);
                                return;
                            }
                        }
                        _ => {
                            error!("Error: unknown key {}", key.as_str());
                            return;
                        }
                    }
                }
            });
        }

        Ok(())
    }

    async fn handle_order_placed(&self, buf: impl Buf) -> Result<()> {
        let order_event = OrderPlacedEvent::decode(buf)?;

        let place_order = match order_event.order {
            Some(v) => v,
            None => return Err(anyhow!("TODO")),
        };

        let mut redis = match self.redis_cl.get_multiplexed_async_connection().await {
            Ok(redis_conn) => redis_conn,
            Err(e) => {
                anyhow::bail!("Failed to estrablish Redis connection: {e}")
            }
        };

        let _: () = redis
            .hset(
                place_order.order_id.clone(),
                "status",
                DeliveryStatus::RiderUnaccept.as_str_name(),
            )
            .await?;

        let (riders, pickup_info) = MyDelivery::prepare_order_delivery(&place_order)
            .context("could not prepare order delivery")?;

        MyDelivery::notify_riders(riders, pickup_info)?;

        self.event_bus
            .publish(
                "rider.notified.event",
                &RiderNotifiedEvent {
                    order_id: place_order.order_id,
                    notify_time: Some(prost_wkt_types::Timestamp::from(Utc::now())),
                }
                .encode_to_vec(),
            )
            .await?;

        Ok(())
    }

    async fn handle_rider_created(&self, buf: impl Buf) -> Result<()> {
        let new_rider = SyncRiderCreated::decode(buf)?;

        let username = new_rider
            .email
            .split_once('@')
            .ok_or(anyhow!("failed to split email to username"))?
            .0
            .to_string();

        self.db
            .create_rider(NewRider {
                rider_id: new_rider.rider_id,
                username,
            })
            .await?;

        Ok(())
    }
}

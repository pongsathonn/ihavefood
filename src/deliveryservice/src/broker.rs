use anyhow::Result;
use lapin::{options::*, types::FieldTable, BasicProperties, Connection, Consumer, ExchangeKind};

#[derive(Debug)]
pub struct RabbitMQ {
    conn: Connection,
}

impl RabbitMQ {
    pub fn new(conn: Connection) -> Self {
        RabbitMQ { conn }
    }

    pub async fn publish(&self, key: &str, payload: &[u8]) -> Result<()> {
        let ch = self.conn.create_channel().await.unwrap();

        ch.exchange_declare(
            "my_exchange",
            ExchangeKind::Direct,
            ExchangeDeclareOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();

        let confirm = ch
            .basic_publish(
                "my_exchange",
                key,
                BasicPublishOptions::default(),
                payload,
                BasicProperties::default(),
            )
            .await
            .unwrap()
            .await
            .unwrap();

        if confirm.is_ack() {
            Ok(())
        } else {
            Err(anyhow::Error::msg("TODO: error"))
        }
    }

    // TODO: handle error properly
    pub async fn subscribe(&self, queue: &str, key: &str) -> Consumer {
        let ch = self.conn.create_channel().await.unwrap();

        ch.exchange_declare(
            "my_exchange",
            ExchangeKind::Direct,
            ExchangeDeclareOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();

        ch.queue_bind(
            queue,
            "my_exchange",
            key,
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();

        ch.basic_consume(
            queue,
            "delivery_service",
            BasicConsumeOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap()
    }
}

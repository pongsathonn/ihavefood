use futures::StreamExt;
use lapin::{options::*, types::FieldTable, Connection};

#[derive(Debug)]
pub struct RabbitMQ {
    conn: Connection,
}

impl RabbitMQ {
    pub fn new(conn: Connection) -> Self {
        RabbitMQ { conn }
    }

    pub async fn publish(&self) {
        todo!();
    }

    pub async fn subscribe(&self, queue: &str, key: &str) {
        let ch = self.conn.create_channel().await.unwrap();

        ch.queue_bind(
            queue,
            "my_exchange",
            key,
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await
        .unwrap();

        let mut consumer = ch
            .basic_consume(
                queue,
                "delivery_service",
                BasicConsumeOptions::default(),
                FieldTable::default(),
            )
            .await
            .unwrap();

        // Read as:
        // while consumer.next`Option<T>` has Some(delivery) then does {}
        // if delivery`Result<T,E>` is Ok(v) then does {}
        while let Some(delivery) = consumer.next().await {
            if let Ok(v) = delivery {
                println!(" [x] Received {:?}", std::str::from_utf8(&v.data).unwrap());
                v.ack(BasicAckOptions::default()).await.unwrap();
            }
        }
    }
}

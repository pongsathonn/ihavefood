use crate::models::*;
use anyhow::Result;
use chrono::offset::Utc;
use chrono::DateTime;
use sqlx::sqlite::SqlitePool;

#[derive(Debug)]
pub struct Db {
    pool: SqlitePool,
}

impl Db {
    pub fn new(pool: SqlitePool) -> Self {
        Db { pool }
    }

    pub async fn create_delivery(&self, new_order: &NewDelivery) -> Result<()> {
        sqlx::query!(
            r#"
        INSERT INTO deliveries (
            order_id, 
            pickup_code,
            pickup_lat,
            pickup_lng,
            drop_off_lat,
            drop_off_lng
        )
        VALUES(?,?,?,?,?,?);
            "#,
            new_order.order_id,
            new_order.pickup_code,
            new_order.pickup_location.latitude,
            new_order.pickup_location.longitude,
            new_order.drop_off_location.latitude,
            new_order.drop_off_location.longitude,
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    pub async fn get_delivery(&self, order_id: String) -> Result<DbDelivery> {
        let foo = sqlx::query!(
            r#"
        SELECT
            deliveries.order_id AS "order_id:String",
            deliveries.rider_id AS "rider_id:i32",
            deliveries.pickup_code AS "pickup_code:String",
            deliveries.pickup_lat AS "pickup_lat:f64",
            deliveries.pickup_lng AS "pickup_lng:f64",
            deliveries.drop_off_lat AS "drop_off_lat:f64",
            deliveries.drop_off_lng AS "drop_off_lng:f64",
            deliveries.status AS "status:DbDeliveryStatus",
            deliveries.create_time AS "create_time:DateTime<Utc>",
            deliveries.accept_time AS "accept_time:DateTime<Utc>",
            deliveries.deliver_time AS "deliver_time:DateTime<Utc>",
            riders.name AS "rider_name:String",
            riders.phone_number AS "rider_phone_number:String"
        FROM deliveries
        LEFT JOIN riders
        ON deliveries.rider_id = riders.id
        WHERE deliveries.order_id = ?
        "#,
            order_id
        )
        .fetch_one(&self.pool)
        .await?;

        Ok(DbDelivery {
            order_id: foo.order_id,
            rider: Some(DbRider {
                id: foo.rider_id,
                name: foo.rider_name,
                phone_number: foo.rider_phone_number,
            }),
            pickup_code: foo.pickup_code,
            pickup_location: DbPoint {
                latitude: foo.pickup_lat,
                longitude: foo.pickup_lng,
            },
            drop_off_location: DbPoint {
                latitude: foo.drop_off_lat,
                longitude: foo.drop_off_lng,
            },
            status: foo.status,
            timestamp: DbTimestamp {
                create_time: foo.create_time,
                accept_time: foo.accept_time,
                deliver_time: foo.deliver_time,
            },
        })
    }

    pub async fn get_delivery_status(&self, order_id: String) -> Result<DbDeliveryStatus> {
        let status = sqlx::query_scalar!(
            r#" SELECT status AS "status!:DbDeliveryStatus"  FROM deliveries WHERE order_id=? "#,
            order_id,
        )
        .fetch_one(&self.pool)
        .await?;
        Ok(status)
    }

    // update the rider whos accepted the order.
    pub async fn update_delivery_rider(&self, update: &DbUpdateDeliveryRider) -> Result<()> {
        sqlx::query!(
            r#"
            UPDATE deliveries
            SET 
                rider_id=?2,
                status=?3,
                accept_time=?4
            WHERE order_id=?1;
            "#,
            update.order_id,
            update.rider_id,
            update.status,
            update.accept_time,
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }
}

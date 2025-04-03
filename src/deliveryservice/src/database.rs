use crate::models::*;
use anyhow::Result;
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
        let status = update.status.to_string();
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
            status,
            update.accept_time,
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }
}

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

    pub async fn get_delivery(&self, order_id: &str) -> Result<DbDelivery> {
        // The query_as! macro doesn't use FromRow, and DbDelivery has a nested struct
        // with conflicting field names, requiring a manual FromRow impl and making fn
        // query_as() necessary.Since sqlx lacks fn query_file_as(), the SQL is loaded
        // using std::fs from the /query directory.
        let sql = std::fs::read_to_string("queries/delivery-by-id.sql")?;
        let delivery = sqlx::query_as(sql.as_str())
            .bind(order_id)
            .fetch_one(&self.pool)
            .await?;

        Ok(delivery)
    }

    pub async fn get_delivery_status(&self, order_id: &str) -> Result<DbDeliveryStatus> {
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

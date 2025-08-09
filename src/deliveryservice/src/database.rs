use crate::models::*;
use anyhow::Result;
use log::info;
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
        sqlx::query_file!(
            "queries/create-delivery.sql",
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

    pub async fn create_rider(&self, new_rider: &NewRider) -> Result<()> {
        sqlx::query_file!(
            "queries/create-rider.sql",
            new_rider.rider_id,
            new_rider.username,
            new_rider.phone_number,
        )
        .execute(&self.pool)
        .await?;
        Ok(())
    }

    pub async fn get_delivery(&self, order_id: &str) -> Result<DbDelivery> {
        // 1) The query_as! macro doesn't use FromRow,
        // 2) DbDelivery has a nested struct with conflicting field names,
        //    requiring a manual FromRow impl and making fn query_as() necessary.
        // 3) Since sqlx lacks fn query_file_as(), the SQL is loaded using std::fs
        //    from the /query directory.
        let sql = std::fs::read_to_string("queries/delivery-by-id.sql")?;
        let delivery = sqlx::query_as(sql.as_str())
            .bind(order_id)
            .fetch_one(&self.pool)
            .await?;
        Ok(delivery)
    }

    pub async fn get_rider(&self, rider_id: &str) -> Result<DbRider> {
        let rider = sqlx::query_file_as!(DbRider, "queries/rider-by-id.sql", rider_id)
            .fetch_one(&self.pool)
            .await?;

        Ok(rider)
    }

    pub async fn get_delivery_status(&self, order_id: &str) -> Result<DbDeliveryStatus> {
        let status = sqlx::query_file_scalar!("queries/delivery-status-by-id.sql", order_id)
            .fetch_one(&self.pool)
            .await?;
        Ok(status)
    }

    // update the rider whos accepted the order.
    pub async fn update_delivery_rider(&self, order_id: &str, rider_id: &str) -> Result<()> {
        sqlx::query_file!("queries/update-delivery-rider.sql", order_id, rider_id,)
            .execute(&self.pool)
            .await?;
        Ok(())
    }

    pub async fn update_delivery_status(
        &self,
        order_id: &str,
        status: DbDeliveryStatus,
    ) -> Result<()> {
        sqlx::query_file!("queries/update-delivery-status.sql", order_id, status,)
            .execute(&self.pool)
            .await?;
        Ok(())
    }
}

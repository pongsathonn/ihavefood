use crate::models::*;
use anyhow::Result;
use anyhow::{anyhow, bail};
use mongodb::{bson::doc, Collection};

#[derive(Debug)]
pub struct Db {
    delivery_coll: Collection<DbDelivery>,
    rider_coll: Collection<DbRider>,
}

impl Db {
    pub fn new(delivery_coll: Collection<DbDelivery>, rider_coll: Collection<DbRider>) -> Self {
        Db {
            delivery_coll,
            rider_coll,
        }
    }

    pub async fn create_rider(&self, new_rider: NewRider) -> Result<()> {
        let _ = self.rider_coll.insert_one(DbRider {
            id: new_rider.rider_id,
            username: new_rider.username,
            phone_number: String::new(),
        });
        Ok(())
    }

    pub async fn get_delivery(&self, order_id: &str) -> Result<DbDelivery> {
        self.delivery_coll
            .find_one(doc! { "order_id":order_id })
            .await?
            .ok_or_else(|| anyhow!(sqlx::Error::RowNotFound))
    }

    pub async fn get_rider(&self, rider_id: &str) -> Result<DbRider> {
        self.rider_coll
            .find_one(doc! { "rider_id":rider_id })
            .await?
            .ok_or_else(|| anyhow!(sqlx::Error::RowNotFound))
    }

    pub async fn update_delivery_rider(&self, order_id: &str, rider_id: &str) -> Result<()> {
        let filter = doc! { "order_id": order_id };
        let update = doc! { "$set": doc! {"rider_id": rider_id} };

        let res = self.delivery_coll.update_one(filter, update).await?;
        if res.modified_count <= 0 {
            bail!("update delivery rider failed")
        }
        return Ok(());
    }
}

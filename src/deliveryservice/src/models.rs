use chrono::{DateTime, Utc};
use sqlx::error::BoxDynError;
use sqlx::sqlite::SqliteValueRef;
use sqlx::Type;
use sqlx::{Decode, Encode, Sqlite};
use std::result::Result;

// dbDelivery represent delivery information for an order
pub struct DbDelivery {
    pub order_id: String,
    // rider_id who accept the order
    pub rider_id: String,
    // Current rider location
    pub rider_location: DbPoint,
    // PickupCode is code 3 digit for rider pickup
    pub pickup_code: String,
    // pickup_location is Restaurant address
    pub pickup_location: DbPoint,
    // drop_off_location is User address
    pub drop_off_location: DbPoint,
    // Delivery status
    pub status: DbDeliveryStatus,
    pub timestamp: DbTimestamp,
}

pub struct DbPoint {
    pub latitude: f64,
    pub longitude: f64,
}

#[derive(PartialEq)]
pub enum DbDeliveryStatus {
    // UNACCEPTED indicates the rider has not yet accepted the order.
    UNACCEPT,
    // ACCEPTED indicates the rider has accepted the order.
    ACCEPTED,
    // DELIVERED indicates the order has been delivered by the rider.
    DELIVERED,
}

impl<'q> Encode<'q, Sqlite> for DbDeliveryStatus {
    fn encode_by_ref(
        &self,
        buf: &mut <Sqlite as sqlx::Database>::ArgumentBuffer<'q>,
    ) -> Result<sqlx::encode::IsNull, BoxDynError> {
        match self {
            DbDeliveryStatus::UNACCEPT => <&str as Encode<'q, Sqlite>>::encode("UNACCEPT", buf),
            DbDeliveryStatus::ACCEPTED => <&str as Encode<'q, Sqlite>>::encode("ACCEPTED", buf),
            DbDeliveryStatus::DELIVERED => <&str as Encode<'q, Sqlite>>::encode("DELIVERED", buf),
        }
    }
}

impl Type<Sqlite> for DbDeliveryStatus {
    fn type_info() -> sqlx::sqlite::SqliteTypeInfo {
        <&str as Type<Sqlite>>::type_info()
    }
}

impl<'r> Decode<'r, Sqlite> for DbDeliveryStatus
where
    &'r str: Decode<'r, Sqlite>,
{
    fn decode(value: SqliteValueRef<'r>) -> Result<Self, BoxDynError> {
        let value = <&str as Decode<Sqlite>>::decode(value)?;

        match value {
            "UNACCEPT" => Ok(DbDeliveryStatus::UNACCEPT),
            "ACCEPTED" => Ok(DbDeliveryStatus::ACCEPTED),
            "DELIVERED" => Ok(DbDeliveryStatus::DELIVERED),
            _ => Err(format!("Invalid delivery status: {}", value).into()),
        }
    }
}

pub struct DbTimestamp {
    // CreateTime is the timestamp when the DeliveryService receives
    // a new order.
    pub create_time: DateTime<Utc>,
    // AcceptTime is the timestamp when the rider accepts the order.
    pub accept_time: DateTime<Utc>,
    // DeliverTime is the timestamp when the order is delivered.
    pub deliver_time: DateTime<Utc>,
}

// NewOrder instead ?
pub struct NewDelivery {
    pub order_id: String,
    pub pickup_code: String,
    pub pickup_location: DbPoint,
    pub drop_off_location: DbPoint,
    pub create_time: DateTime<Utc>,
}

pub struct DbUpdateDeliveryRider {
    pub order_id: String,
    pub rider_id: String,
    pub accept_time: DateTime<Utc>,
    pub status: DbDeliveryStatus,
}

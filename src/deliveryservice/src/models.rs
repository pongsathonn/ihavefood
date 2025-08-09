use chrono::{DateTime, Utc};
use sqlx::{sqlite::SqliteRow, FromRow, Row};

// DbDelivery represent delivery record information for an order
pub struct DbDelivery {
    pub order_id: String,
    // rider who accept the order
    pub rider: Option<DbRider>,
    // PickupCode is code 3 digit for rider pickup
    pub pickup_code: String,
    // pickup_location is Merchant address
    pub pickup_location: DbPoint,
    // drop_off_location is User address
    pub drop_off_location: DbPoint,
    // Delivery status
    pub status: DbDeliveryStatus,
    pub timestamp: DbTimestamp,
}

#[derive(Debug, FromRow)]
pub struct DbRider {
    pub id: String,
    pub username: String,
    pub phone_number: String,
}

#[derive(Debug, PartialEq, sqlx::Type)]
#[repr(i32)]
pub enum DbDeliveryStatus {
    // UNACCEPTED indicates the rider has not yet accepted the order.
    Unaccept,
    // ACCEPTED indicates the rider has accepted the order.
    Accepted,
    // DELIVERED indicates the order has been delivered by the rider.
    Delivered,
}

pub struct DbTimestamp {
    // CreateTime is the timestamp when the DeliveryService receives
    // a new order.
    pub create_time: DateTime<Utc>,
    // AcceptTime is the timestamp when the rider accepts the order.
    pub accept_time: Option<DateTime<Utc>>,
    // DeliverTime is the timestamp when the order is delivered.
    pub deliver_time: Option<DateTime<Utc>>,
}

// use this if want to query only point
pub struct DbPoint {
    pub latitude: f64,
    pub longitude: f64,
}

// NewOrder instead ?
pub struct NewDelivery {
    pub order_id: String,
    pub pickup_code: String,
    pub pickup_location: DbPoint,
    pub drop_off_location: DbPoint,
    pub create_time: DateTime<Utc>,
}

pub struct NewRider {
    pub rider_id: String,
    pub username: String,
    pub phone_number: String,
}

//==============================
//            impl
//==============================

impl FromRow<'_, SqliteRow> for DbDelivery {
    fn from_row(row: &SqliteRow) -> sqlx::Result<Self> {
        Ok(Self {
            order_id: row.try_get("order_id")?,
            rider: Some(DbRider {
                id: row.try_get("rider_id")?,
                username: row.try_get("rider_name")?,
                phone_number: row.try_get("rider_phone_number")?,
            }),
            pickup_code: row.try_get("pickup_code")?,
            pickup_location: DbPoint {
                latitude: row.try_get("pickup_lat")?,
                longitude: row.try_get("pickup_lng")?,
            },
            drop_off_location: DbPoint {
                latitude: row.try_get("drop_off_lat")?,
                longitude: row.try_get("drop_off_lng")?,
            },
            status: row.try_get("status")?,
            timestamp: DbTimestamp {
                create_time: row.try_get("create_time")?,
                accept_time: row.try_get("accept_time")?,
                deliver_time: row.try_get("deliver_time")?,
            },
        })
    }
}

// impl FromRow<'_, SqliteRow> for DbRider {
//     fn from_row(row: &SqliteRow) -> sqlx::Result<Self> {
//         Ok(Self {
//             id: row.try_get("id")?,
//             username: row.try_get("username")?,
//             phone_number: row.try_get("phone_number")?,
//         })
//     }
// }

//impl<'q> Encode<'q, Sqlite> for DbDeliveryStatus {
//    fn encode_by_ref(
//        &self,
//        buf: &mut <Sqlite as sqlx::Database>::ArgumentBuffer<'q>,
//    ) -> Result<sqlx::encode::IsNull, BoxDynError> {
//        match self {
//            DbDeliveryStatus::UNACCEPT => <&str as Encode<'q, Sqlite>>::encode("UNACCEPT", buf),
//            DbDeliveryStatus::ACCEPTED => <&str as Encode<'q, Sqlite>>::encode("ACCEPTED", buf),
//            DbDeliveryStatus::DELIVERED => <&str as Encode<'q, Sqlite>>::encode("DELIVERED", buf),
//        }
//    }
//}
//
//impl Type<Sqlite> for DbDeliveryStatus {
//    fn type_info() -> sqlx::sqlite::SqliteTypeInfo {
//        <&str as Type<Sqlite>>::type_info()
//    }
//}
//
//impl<'r> Decode<'r, Sqlite> for DbDeliveryStatus
//where
//    &'r str: Decode<'r, Sqlite>,
//{
//    fn decode(value: SqliteValueRef<'r>) -> Result<Self, BoxDynError> {
//        let value = <&str as Decode<Sqlite>>::decode(value)?;
//
//        match value {
//            "UNACCEPT" => Ok(DbDeliveryStatus::UNACCEPT),
//            "ACCEPTED" => Ok(DbDeliveryStatus::ACCEPTED),
//            "DELIVERED" => Ok(DbDeliveryStatus::DELIVERED),
//            _ => Err(format!("Invalid delivery status: {}", value).into()),
//        }
//    }
//}

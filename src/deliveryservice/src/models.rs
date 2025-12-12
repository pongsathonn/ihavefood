use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

// NewOrder instead ?
// pub struct NewDelivery {
//     pub order_id: String,
//     pub pickup_code: String,
//     pub pickup_location: DbPoint,
//     pub drop_off_location: DbPoint,
//     pub create_time: DateTime<Utc>,
// }

#[derive(Serialize, Deserialize, Debug)]
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
    // pub status: DeliveryStatus,
    pub timestamp: DbTimestamp,
}

pub struct NewRider {
    pub rider_id: String,
    pub username: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct DbRider {
    pub id: String,
    pub username: String,
    pub phone_number: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct DbPoint {
    pub latitude: f64,
    pub longitude: f64,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct DbTimestamp {
    // CreateTime is the timestamp when the DeliveryService receives
    // a new order.
    pub create_time: DateTime<Utc>,
    // AcceptTime is the timestamp when the rider accepts the order.
    pub accept_time: Option<DateTime<Utc>>,
    // DeliverTime is the timestamp when the order is delivered.
    pub deliver_time: Option<DateTime<Utc>>,
}

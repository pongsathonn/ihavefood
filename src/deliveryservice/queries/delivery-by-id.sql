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
LEFT JOIN riders ON deliveries.rider_id = riders.id
WHERE deliveries.order_id = ?





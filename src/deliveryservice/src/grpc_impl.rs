use crate::delivery::MyDelivery;
use crate::ihavefood::delivery_service_server::DeliveryService;
use crate::ihavefood::*;
use crate::models::{DbDeliveryStatus, NewRider};

use log::error;
use rand::Rng;
use tokio::sync::mpsc;
use tokio::time::{sleep, Duration};
use tokio_stream::wrappers::ReceiverStream;
use tonic::{Code, Request, Response, Status};

#[tonic::async_trait]
impl DeliveryService for MyDelivery {
    type GetOrderTrackingStream = ReceiverStream<Result<GetOrderTrackingResponse, Status>>;

    async fn get_order_tracking(
        &self,
        request: Request<GetOrderTrackingRequest>,
    ) -> Result<Response<Self::GetOrderTrackingStream>, Status> {
        let _ = request;
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            for _ in 1..6 {
                sleep(Duration::from_secs(5)).await;

                // TODO: tracking rider location from GoogleAPI or database
                if tx
                    .send(Ok(GetOrderTrackingResponse {
                        ..Default::default()
                    }))
                    .await
                    .is_err()
                {
                    error!("receiver dropped");
                    return;
                }
            }
        });

        Ok(Response::new(ReceiverStream::new(rx)))
    }

    async fn get_delivery_fee(
        &self,
        request: Request<GetDeliveryFeeRequest>,
    ) -> Result<Response<GetDeliveryFeeResponse>, Status> {
        let customer = self
            .customercl
            .clone()
            .get_customer(GetCustomerRequest {
                customer_id: request.get_ref().customer_id.clone(),
            })
            .await?;
        let customer_addr = customer
            .get_ref()
            .addresses
            .iter()
            .find(|&addr| addr.address_id == request.get_ref().customer_address_id)
            .ok_or(Status::new(Code::Internal, "internal server error"))?;

        let merchant = self
            .merchantcl
            .clone()
            .get_merchant(GetMerchantRequest {
                merchant_id: request.get_ref().merchant_id.clone(),
            })
            .await?;

        let merchant_addr = merchant
            .get_ref()
            .address
            .as_ref()
            .ok_or(Status::new(Code::Internal, "internal server error"))?;

        let customer_point = fake_geocode(customer_addr);
        let merchant_point = fake_geocode(merchant_addr);
        let fee = Self::calc_delivery_fee(&customer_point, &merchant_point).map_err(|err| {
            error!("calculate delivery fee: {err}");
            Status::new(Code::Internal, "failed to calculate delivery fee")
        })?;

        Ok(Response::new(GetDeliveryFeeResponse { fee }))
    }

    async fn confirm_rider_accept(
        &self,
        request: Request<ConfirmRiderAcceptRequest>,
    ) -> Result<Response<PickupInfo>, Status> {
        let delivery = match self
            .db
            .get_delivery(request.get_ref().order_id.as_str())
            .await
        {
            Ok(v) => v,
            Err(e) => {
                error!("database get delivery:{}", e);
                return Err(Status::internal("Failed to get delivery"));
            }
        };

        match delivery.status {
            DbDeliveryStatus::Unaccept => (),
            DbDeliveryStatus::Delivered => {
                return Err(Status::invalid_argument("rider already accepted"))
            }
            DbDeliveryStatus::Accepted => {
                return Err(Status::invalid_argument("order already delivered"))
            }
        }

        // TODO: push notify rider has accepted the order

        self.db
            .update_delivery_rider(
                request.get_ref().order_id.as_str(),
                request.get_ref().rider_id.as_str(),
            )
            .await
            .unwrap();

        self.db
            .update_delivery_status(
                request.get_ref().order_id.as_str(),
                DbDeliveryStatus::Accepted,
            )
            .await
            .unwrap();

        Ok(Response::new(PickupInfo {
            pickup_code: delivery.pickup_code,
            pickup_location: Some(Point {
                latitude: delivery.pickup_location.latitude,
                longitude: delivery.pickup_location.longitude,
            }),
            drop_off_location: Some(Point {
                latitude: delivery.drop_off_location.latitude,
                longitude: delivery.drop_off_location.longitude,
            }),
        }))
    }

    async fn confirm_order_deliver(
        &self,
        request: Request<ConfirmOrderDeliverRequest>,
    ) -> Result<Response<::prost_wkt_types::Empty>, Status> {
        self.db
            .update_delivery_status(
                request.into_inner().order_id.as_str(),
                DbDeliveryStatus::Delivered,
            )
            .await
            .unwrap();
        Ok(Response::new(::prost_wkt_types::Empty {}))
    }

    async fn create_rider(
        &self,
        request: Request<CreateRiderRequest>,
    ) -> Result<Response<Rider>, Status> {
        let req = request.into_inner();
        let new_rider = &NewRider {
            rider_id: req.rider_id.clone(),
            username: req.username,
            phone_number: String::from(""),
        };

        let _ = self.db.create_rider(new_rider).await.unwrap();
        let rider = self.db.get_rider(req.rider_id.as_str()).await.unwrap();

        Ok(Response::new(Rider {
            rider_id: rider.id,
            username: rider.username,
            phone_number: rider.phone_number,
        }))
    }
}

// fake_geocode from ChatGPT
pub fn fake_geocode(_addr: &Address) -> Point {
    let mut rng = rand::rng();

    // arbitrary “center” at 0,0 and offset within ~25 km (~0.225 lat, ~0.25 lng)
    let max_lat_offset = 0.225;
    let max_lng_offset = 0.25;

    Point {
        latitude: rng.random_range(-max_lat_offset..=max_lat_offset),
        longitude: rng.random_range(-max_lng_offset..=max_lng_offset),
    }
}

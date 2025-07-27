use crate::delivery::MyDelivery;
use crate::ihavefood::delivery_service_server::DeliveryService;
use crate::ihavefood::*;
use crate::models::DbDeliveryStatus;

use log::error;
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
        let restaurant_point = Point {
            latitude: request.get_ref().restaurant_lat,
            longitude: request.get_ref().restaurant_long,
        };

        let customer_point = Point {
            latitude: request.get_ref().customer_lat,
            longitude: request.get_ref().customer_long,
        };

        let delivery_fee =
            Self::calc_delivery_fee(&customer_point, &restaurant_point).map_err(|err| {
                error!("Error: {err}");
                Status::new(Code::Internal, "failed to calculate delivery fee")
            })?;

        Ok(Response::new(GetDeliveryFeeResponse { delivery_fee }))
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
                error!("Failed to query:{}", e);
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
        unimplemented!();
    }
}

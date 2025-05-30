use std::pin::Pin;

use crate::proto::v1;
use tokio::sync::mpsc;
use tokio_stream::{Stream, wrappers::ReceiverStream};
use tonic::{Request, Response, Status, Streaming};

pub struct TransformService {}

#[tonic::async_trait]
impl v1::transform_service_server::TransformService for TransformService {
    type TransformStreamStream =
        Pin<Box<dyn Stream<Item = Result<v1::TransformResponse, Status>> + Send>>;

    async fn transform_stream(
        &self,
        _req: Request<Streaming<v1::TransformRequest>>,
    ) -> tonic::Result<tonic::Response<Self::TransformStreamStream>> {
        let (_tx, rx) = mpsc::channel(128);
        Ok(Response::new(Box::pin(ReceiverStream::new(rx))))
    }
}

impl TransformService {
    pub fn new() -> Self {
        Self {}
    }
}

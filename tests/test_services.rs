use avf_server::{
    proto::v1::transform_service_server::TransformService as _, service::TransformService,
};

mod mock;

lazy_static::lazy_static! {
    pub static ref SERVICE: TransformService = TransformService::new() ;
}

#[tokio::test]
async fn test() {
    let request = mock::request([]);
    let _m = SERVICE.transform_stream(request).await.unwrap();
}

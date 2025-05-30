#![warn(clippy::pedantic)]

use std::{net::SocketAddr, str::FromStr};

use clap::Parser;
use proto::v1;
use tonic::{codec::CompressionEncoding, transport::Server};

mod proto;
mod service;

#[derive(Parser)]
#[command(version, about, long_about = None)]
/// AVF-Server
struct Cmd {
    /// listen address
    #[arg(short, long, default_value_t = SocketAddr::from_str("[::]:4000").unwrap())]
    listen: SocketAddr,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt::init();
    let args = Cmd::parse();

    let addr = args.listen;
    tracing::info!("Listening on {}", addr);
    Server::builder()
        .add_service(
            tonic_reflection::server::Builder::configure()
                .register_encoded_file_descriptor_set(proto::FILE_DESCRIPTOR_SET)
                .build_v1()
                .unwrap(),
        )
        .add_service(
            tonic_reflection::server::Builder::configure()
                .register_encoded_file_descriptor_set(proto::FILE_DESCRIPTOR_SET)
                .build_v1alpha()
                .unwrap(),
        )
        .add_service(
            v1::transform_service_server::TransformServiceServer::new(
                service::TransformService::new(),
            )
            .send_compressed(CompressionEncoding::Gzip)
            .accept_compressed(CompressionEncoding::Gzip),
        )
        .serve(addr)
        .await?;

    Ok(())
}

#![allow(clippy::pedantic)]

pub mod v1 {
    tonic::include_proto!("avf.v1");
}

pub const FILE_DESCRIPTOR_SET: &[u8] = tonic::include_file_descriptor_set!("avf_descriptor");

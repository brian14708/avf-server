use std::{env, path::PathBuf};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());

    tonic_build::configure()
        .build_client(false)
        .file_descriptor_set_path(out_dir.join("avf_descriptor.bin"))
        .compile_protos(&["proto/avf/v1/transform.proto"], &["proto"])?;
    Ok(())
}

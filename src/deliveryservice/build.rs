use std::{env, error::Error, fs};

fn main() -> Result<(), Box<dyn Error>> {
    // The protobuf files are located locally using relative paths.
    // When building in Docker, this can cause issues.
    // To avoid this, ensure this build runs only on the local.
    //
    // NOTE: variable in .cargo/config.yaml can be overrided with shell variable.
    //
    // WKT = standard protobuf types from `google.protobuf` like Timestamp.
    // see generate wkt issues on: https://github.com/tokio-rs/prost/issues/672
    if env!("PLATFORM") == "host" {
        // Use /genproto as the output dir for consistency across services.
        let _ = fs::create_dir("genproto");

        tonic_build::configure()
            .out_dir("genproto")
            .type_attribute(".", "#[derive(serde::Deserialize, serde::Serialize)]")
            .compile_well_known_types(true)
            .extern_path(".google.protobuf.Timestamp", "::prost_wkt_types::Timestamp")
            .extern_path(".google.protobuf.Empty", "::prost_wkt_types::Empty")
            .compile_protos(
                &[
                    "../../protos/deliveryservice.proto",
                    "../../protos/merchantservice.proto",
                    "../../protos/customerservice.proto",
                    "../../protos/events.proto",
                ],
                &["../../protos"],
            )?;
    }

    Ok(())
}

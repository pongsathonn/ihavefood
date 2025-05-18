use std::{env, error::Error, fs::create_dir};

fn main() -> Result<(), Box<dyn Error>> {
    let _ = create_dir("genproto");

    // The protobuf files are located locally using relative paths.
    // When building in Docker, this can cause issues.
    // To avoid this, ensure this build runs only on the local.
    //
    // NOTE: variable in .cargo/config.yaml can be overrided with shell variable.
    if env!("PLATFORM") == "host" {
        tonic_build::configure()
            .out_dir("genproto")
            .compile_protos(
                &[
                    "../../protos/orderservice.proto",
                    "../../protos/deliveryservice.proto",
                ],
                &["../../protos"],
            )?;
    }
    Ok(())
}

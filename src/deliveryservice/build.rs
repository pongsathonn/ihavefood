fn main() -> Result<(), Box<dyn std::error::Error>> {
    // NOTE:
    // - set at $CARGO_HOME/config.toml
    // - ensure /genproto dir is created.
    if let Ok(v) = std::env::var("BUILD_MACHINE") {
        if v == "local" {
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
    }

    Ok(())
}

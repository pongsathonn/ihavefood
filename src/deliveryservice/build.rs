fn main() -> Result<(), Box<dyn std::error::Error>> {
    // NOTE:
    // - set at $CARGO_HOME/config.toml
    // - ensure /genproto dir is created.
    if std::env::var("BUILD_MACHINE")? == "local" {
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

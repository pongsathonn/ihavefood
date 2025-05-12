fn main() -> Result<(), Box<dyn std::error::Error>> {
    // NOTE:
    // - set at $CARGO_HOME/config.toml
    // - ensure /genproto dir is created.
<<<<<<< HEAD
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
=======
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
>>>>>>> 662b2561f3aae9c11e203a2b92997f242dc19b49
    }

    Ok(())
}

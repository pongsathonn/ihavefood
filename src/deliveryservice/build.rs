use tonic_build::*;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    configure()
        .type_attribute(
            "PlaceOrder",
            "#[derive(serde::Deserialize, serde::Serialize)]",
        )
        .type_attribute("Menu", "#[derive(serde::Deserialize, serde::Serialize)]")
        .type_attribute("Address", "#[derive(serde::Deserialize, serde::Serialize)]")
        .type_attribute(
            "ContactInfo",
            "#[derive(serde::Deserialize, serde::Serialize)]",
        )
        .type_attribute(
            "OrderTimestamps",
            "#[derive(serde::Deserialize, serde::Serialize)]",
        )
        .compile_protos(
            &[
                "../../protos/orderservice.proto",
                "../../protos/deliveryservice.proto",
            ],
            &["../../protos"],
        )?;

    Ok(())
}

// # use tonic_build::*;
// let mut attributes = Attributes::default();
// attributes.push_struct("EchoService", "#[derive(PartialEq)]");

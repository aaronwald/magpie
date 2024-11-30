use paho_mqtt as mqtt;
use std::{env, process, time::Duration};
use log::{info, error, warn};
use serde_json::Value;
use futures::stream::select_all;
use futures::{executor::block_on, stream::StreamExt};

const SUBS: &[&str] = &["rtl_433/Acurite-Atlas/622/msg5"];
const QOS: &[i32] = &[mqtt::QOS_0];

// TODO tokio https://docs.rs/rumqttc/latest/rumqttc/
// TODO ctrl+c (with tokio?)
fn main() {
    let _ec = |x: i32| -> i32 {
        let yx = x + 1;
        yx
    };
    
    // Initialize the logger from the environment
    env_logger::init();
    
    let host = env::args()
    .nth(1)
    .unwrap_or_else(|| "mqtt://homeassistant.local:1883".to_string());
    
    info!("Connecting to the MQTT broker at '{}'...", host);
    
    // Create the client. Use an ID for a persistent session.
    // A real system should try harder to use a unique ID.
    let create_opts = mqtt::CreateOptionsBuilder::new_v3()
    .server_uri(host)
    .client_id("rusty-async-subscriber")
    .finalize();
    
    // Create the client connection
    let mut cli = mqtt::AsyncClient::new(create_opts).unwrap_or_else(|e| {
        error!("Error creating the client: {:?}", e);
        process::exit(1);
    });
    
    if let Err(err) = block_on(async {
        let strm = cli.get_stream(25);
        
        // Define the set of options for the connection
        let _lwt = mqtt::Message::new(
            "test/lwt",
            "[LWT] Async subscriber lost connection",
            mqtt::QOS_1,
        );
        
        // Create the connect options, explicitly requesting MQTT v3.x
        let conn_opts = mqtt::ConnectOptionsBuilder::new_v3()
        .keep_alive_interval(Duration::from_secs(30))
        .clean_session(false)
        // .will_message(lwt)
        .user_name("mqtt")
        .password("mqtt")
        .finalize();
        
        // Make the connection to the broker
        cli.connect(conn_opts).await?;
        
        info!("Subscribing to topics: {:?}", SUBS);
        cli.subscribe_many(SUBS, QOS).await?;
        
        let mut rconn_attempt: usize = 0;
        let streams: Vec<_> = vec![strm];
        
        let mut fused_streams = select_all(streams);
        while let Some(value) = fused_streams.next().await {
            if let Some(msg) = value {
                let s: String = msg.payload_str().into_owned();
                info!("Received message: {:?}", msg.topic());
                serde_json::from_str(&s).map(|json: Value| {
                    json.get("humidity").map(|temp| {
                        info!("humidity: {:?}", temp);
                    }).unwrap_or_else(|| {
                        error!("humidity not found");
                    });
                    
                    json.get("temperature_F").map(|temp| {
                        info!("temp(F): {:?}", temp);
                    });
                }).unwrap_or_else(|e| {
                    error!("Error parsing JSON: {:?}", e);
                });
            } else {
                // A "None" means we were disconnected. Try to reconnect...
                warn!("Lost connection. Attempting reconnect...");
                while let Err(err) = cli.reconnect().await {
                    rconn_attempt += 1;
                    warn!("Error reconnecting #{}: {}", rconn_attempt, err);
                    // For tokio use: tokio::time::delay_for()
                    async_std::task::sleep(Duration::from_secs(1)).await;
                }
                warn!("Reconnected.");
            }
        }
        
        // Explicit return type for the async block
        Ok::<(), mqtt::Error>(())
    }) {
        warn!("{}", err);
    }
}
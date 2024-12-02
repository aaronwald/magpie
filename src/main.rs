use tokio::net::{TcpListener, TcpStream};
use tokio_tungstenite::{
    accept_async,
    tungstenite::{Error, Result},
};
use tokio_tungstenite::tungstenite::protocol::Message;
use futures_util::{StreamExt, SinkExt};
use log::*;
use std::net::SocketAddr;
use tokio::sync::mpsc;
use tokio::sync::mpsc::Sender;
use std::{process, time::Duration};
use paho_mqtt as mqtt;
use serde_json::Value;
const QOS: &[i32] = &[mqtt::QOS_0];
use futures::stream::select_all;

const SUBS: &[&str] = &["rtl_433/Acurite-Atlas/622/msg5"];

async fn domqtt (ctx: Sender<String>) -> Result<()> {
    let host = "mqtt://homeassistant.local:1883".to_string();

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

    let conn_opts = mqtt::ConnectOptionsBuilder::new_v3()
    .keep_alive_interval(Duration::from_secs(30))
    .clean_session(false)
    // .will_message(lwt)
    .user_name("mqtt")
    .password("mqtt")
    .finalize();
    cli.connect(conn_opts).await.unwrap_or_else(|e| {
        error!("Unable to connect: {:?}", e);
        process::exit(1);
    });

    cli.subscribe_many(SUBS, QOS).await.unwrap_or_else(|e| {
        error!("Error subscribing to topics: {:?}", e);
        process::exit(1);
    });
    let strm = cli.get_stream(25);
    let streams: Vec<_> = vec![strm];
        
        
    let mut fused_streams = select_all(streams);
    while let Some(value) = fused_streams.next().await {
        if let Some(msg) = value {
            let s: String = msg.payload_str().into_owned();
            info!("Received message: {:?}", msg.topic());

            if let Ok(json) = serde_json::from_str::<Value>(&s) {
                if let Some(humidity) = json.get("humidity") {
                    info!("humidity: {:?}", humidity);
                } else {
                    error!("humidity not found");
                }

                if let Some(temp) = json.get("temperature_F") {
                    let temp_str = format!("{:?}", temp);
                    info!("temp(F): {:?}", temp_str);
                    if !temp_str.is_empty() {
                        let json_temp = serde_json::json!({ "temperature_F": temp });
                        ctx.send(json_temp.to_string()).await.expect("Error sending message");
                    }
                } else {
                    error!("temperature_F not found");
                }
            } else {
                error!("Error parsing JSON: {:?}", s);
            }

            // ctx.send(s.to_string()).await.expect("Error sending message");
        } else {
            warn!("Lost connection. Attempting reconnect...");
            cli.reconnect().await.unwrap_or_else(|e| {
                error!("Error reconnecting: {:?}", e);
                process::exit(1);
            });
        }
    }

    Ok(())
}

async fn accept_connection(peer: SocketAddr, stream: TcpStream) {
    if let Err(e) = handle_connection(peer, stream).await {
        match e {
            Error::ConnectionClosed | Error::Protocol(_) | Error::Utf8 => (),
            err => error!("Error processing connection: {}", err),
        }
    }
}

async fn handle_connection(peer: SocketAddr, stream: TcpStream) -> Result<()> {
    let ws_stream = accept_async(stream).await.expect("Error during the websocket handshake");

    info!("New WebSocket connection: {}", peer);

    let (mut write, mut read) = ws_stream.split();
    let (ctx, mut crx) = mpsc::channel(32);
    let cctx = ctx.clone();

    tokio::spawn(async move {
        while let Some(message) = crx.recv().await {
            write.send(Message::Text(message)).await.expect("Error sending message");
            // TODO how to close channel?
        }
    });

    tokio::spawn(async move {
        let _ = domqtt(cctx).await;
    });

    while let Some(message) = read.next().await {
        let message = message.expect("Error reading message");
        if message.is_text() || message.is_binary() {
            let s = format!("{msg}", msg = message.to_string());
            info!("Received message: {}", s);
        }
    }
    Ok(())
}

#[tokio::main]
async fn main() {
    env_logger::init();

    let addr = "0.0.0.0:8080";
    let listener = TcpListener::bind(&addr).await.expect("Failed to bind");

    println!("Listening on: {}", addr);

    while let Ok((stream, _)) = listener.accept().await {
        let peer = stream.peer_addr().expect("connected streams should have a peer address");
        info!("Peer address: {}", peer);

        tokio::spawn(accept_connection(peer, stream));
    }
}
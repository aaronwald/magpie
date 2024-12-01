
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
use futures::{executor::block_on};

const SUBS: &[&str] = &["rtl_433/Acurite-Atlas/622/msg5"];

async fn domqtt (ctx: Sender<String>) -> Result<()> {
    ctx.send("domqtt".to_string()).await.expect("Error sending message");
    
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
                        ctx.send(format!("temp(F): {:?}", temp));
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

    // spawn a task that reads a bounded channel and writes to the websocket
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(std::time::Duration::from_secs(5));
        loop {
            interval.tick().await;
            ctx.send("ping".to_string()).await.expect("Error sending message");
        }
    });

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
            // just log
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
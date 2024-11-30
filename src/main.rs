use tokio::net::{TcpListener, TcpStream};
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::protocol::Message;
use futures_util::{StreamExt, SinkExt};

async fn handle_connection(stream: TcpStream) {
    let ws_stream = accept_async(stream).await.expect("Error during the websocket handshake");

    let (mut write, mut read) = ws_stream.split();

    while let Some(message) = read.next().await {
        let message = message.expect("Error reading message");
        if message.is_text() || message.is_binary() {
            print!("Received message: {}",  message.to_string().as_str());
            write.send(Message::Text("Hello from server".to_string())).await.expect("Error sending message");
        }
    }
}

#[tokio::main]
async fn main() {
    let addr = "0.0.0.0:8080";
    let listener = TcpListener::bind(&addr).await.expect("Failed to bind");

    println!("Listening on: {}", addr);

    while let Ok((stream, _)) = listener.accept().await {
        tokio::spawn(handle_connection(stream));
    }
}
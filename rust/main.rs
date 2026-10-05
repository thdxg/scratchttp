use anyhow::{Context, Result};
use std::{
    net::{TcpListener, TcpStream},
    thread,
};

fn main() -> Result<()> {
    let port = std::env::var("PORT")
        .context("PORT not set")?
        .parse::<u16>()
        .context("failed to parse port")?;

    let addr = format!("localhost:{port}");

    let listener = TcpListener::bind(addr)?;

    for conn in listener.incoming() {
        match conn {
            Ok(stream) => {
                thread::spawn(move || handle_client(stream));
            }
            Err(err) => return Err(err.into()),
        }
    }

    Ok(())
}

fn handle_client(stream: TcpStream) -> Result<()> {
    Ok(())
}

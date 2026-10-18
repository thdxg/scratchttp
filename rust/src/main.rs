use crate::server::Server;
use anyhow::{Context, Result};
use std::env;

mod http;
mod server;

fn main() -> Result<()> {
    let addr = env::var("ADDRESS").context("ADDRESS not set")?;

    let srv = Server::new(addr.as_str());

    srv.run()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_main() {
        let path = env::var("TESTDATA").context("TESTDATA not set").unwrap();

    }
}

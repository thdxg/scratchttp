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
    use std::{fs::File, thread, time::Duration};

    #[test]
    fn test_main() -> Result<()> {
        let addr = env::var("ADDRESS").context("ADDRESS not set")?;
        let path = env::var("TESTDATA").context("TESTDATA not set")?;
        let file = File::open(path).context("failed to open testdata")?;
        let data: Vec<String> = serde_json::from_reader(file).context("failed to read testdata")?;

        thread::spawn(main);
        thread::sleep(Duration::from_secs(1)); // wait for server

        let c = reqwest::blocking::Client::new();
        let url = format!("http://{addr}/store");

        thread::scope(|s| {
            let c = &c;
            let url = &url;
            for d in &data {
                s.spawn(move || {
                    let res = c
                        .post(url)
                        .header("Content-Type", "text/plain")
                        .body(d.clone())
                        .send()
                        .expect("request failed");

                    assert_eq!(res.status(), 200);
                });
            }
        });

        let mut data = data;
        data.sort();

        let res = c.get(url).send().context("request failed")?;
        assert_eq!(
            res.json::<Vec<String>>()
                .context("failed to read response")?,
            data
        );

        Ok(())
    }
}

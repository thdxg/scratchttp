use anyhow::{Context, Result};
use std::{collections::HashMap, net::TcpStream};
use std::{
    net::TcpListener,
    thread::{self},
};
use crate::http::{self, Header};
use super::db::DB;

pub struct Server {
    pub addr: String,
    pub db: DB,
    pub routes: HashMap<&'static str, Handler>,
}

type Handler = fn(&Server, &http::Request) -> http::Response;

impl Server {
    pub fn new(addr: &str) -> Self {
        let addr = addr.to_owned();
        let routes = HashMap::from([
            ("/", Self::handle_index as Handler),
            ("/echo", Self::handle_echo as Handler),
            ("/store", Self::handle_store as Handler),
        ]);
        let db = DB::new();
        Server { addr, routes, db }
    }

    pub fn run(&self) -> Result<()> {
        let listener = TcpListener::bind(&self.addr).context("failed to bind tcp")?;

        thread::scope(|s| {
            for conn in listener.incoming() {
                match conn {
                    Ok(conn) => s.spawn(move || {
                        match self.handle_conn(conn) {
                            Ok(_) => (),
                            Err(e) => eprintln!("failed to handle connection: {e}"),
                        };
                    }),
                    Err(e) => {
                        eprintln!("failed to accept connection: {e}");
                        continue;
                    }
                };
            }
        });

        Ok(())
    }

    fn handle_conn(&self, mut conn: TcpStream) -> Result<()> {
        let req = http::read_request(&mut conn)?;
        let res = self.handle_request(&req);
        http::write_response(&mut conn, &res)?;

        println!("{} {} {}", res.status_code, req.method, req.uri);

        Ok(())
    }

    fn handle_request(&self, req: &http::Request) -> http::Response {
        let mut res = if req.parse_result.is_err() {
            http::Response::with_status(400, "Bad Request")
        } else if let Some(handler) = self.routes.get(req.uri.as_str()) {
            handler(self, req)
        } else {
            http::Response::with_status(404, "Not Found")
        };

        res.headers.push(Header {
            name: "Location".to_string(),
            value: format!("{}{}", req.host, req.uri),
        });
        res.headers.push(Header {
            name: "Content-Length".to_string(),
            value: res.body.len().to_string(),
        });
        res.headers.push(Header {
            name: "Connection".to_string(),
            value: "Close".to_string(),
        });

        res
    }
}

impl Default for http::Response {
    fn default() -> Self {
        Self {
            http_version: "HTTP/1.1",
            status_code: 200,
            reason_pharse: "OK",
            headers: Vec::default(),
            body: String::default(),
        }
    }
}

impl http::Response {
    fn with_status(status_code: u16, reason_pharse: &'static str) -> Self {
        Self {
            status_code,
            reason_pharse,
            ..Default::default()
        }
    }
}

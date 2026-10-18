use crate::{http, server::Server};

impl Server {
    pub fn handle_index(&self, req: &http::Request) -> http::Response {
        let mut res = match req.method {
            http::HttpMethod::Get => http::Response {
                status_code: 200,
                reason_pharse: "OK",
                body: "hi".to_string(),
                ..http::Response::default()
            },
            _ => http::Response {
                status_code: 405,
                reason_pharse: "Method Not Allowed",
                ..http::Response::default()
            },
        };

        res.headers.push(http::Header {
            name: "Content-Type".to_string(),
            value: "text/plain".to_string(),
        });

        res
    }

    pub fn handle_echo(&self, req: &http::Request) -> http::Response {
        let mut res = match req.method {
            http::HttpMethod::Post => http::Response {
                status_code: 200,
                reason_pharse: "OK",
                body: req.body.clone(),
                ..http::Response::default()
            },
            _ => http::Response {
                status_code: 405,
                reason_pharse: "Method Not Allowed",
                ..http::Response::default()
            },
        };

        res.headers.push(http::Header {
            name: "Content-Type".to_string(),
            value: "text/plain".to_string(),
        });

        res
    }

    pub fn handle_store(&self, req: &http::Request) -> http::Response {
        let mut res = match req.method {
            http::HttpMethod::Get => {
                let result = self
                    .db
                    .get()
                    .and_then(|data| Ok(serde_json::to_string(&data)?));
                match result {
                    Ok(body) => http::Response {
                        body,
                        ..http::Response::default()
                    },
                    Err(_) => http::Response {
                        status_code: 500,
                        reason_pharse: "Internal Server Error",
                        ..http::Response::default()
                    },
                }
            }
            http::HttpMethod::Post => {
                let result = self
                    .db
                    .add(req.body.clone())
                    .and_then(|_| self.db.get())
                    .and_then(|data| Ok(serde_json::to_string(&data)?));

                match result {
                    Ok(body) => http::Response {
                        body,
                        ..http::Response::default()
                    },
                    Err(_) => http::Response {
                        status_code: 500,
                        reason_pharse: "Internal Server Error",
                        ..http::Response::default()
                    },
                }
            }
            _ => http::Response {
                status_code: 405,
                reason_pharse: "Method Not Allowed",
                ..http::Response::default()
            },
        };

        res.headers.push(http::Header {
            name: "Content-Type".to_string(),
            value: "text/plain".to_string(),
        });

        res
    }
}

use super::Header;
use anyhow::{Context, Result, anyhow, bail};
use std::{
    fmt::{self},
    io::{BufRead, BufReader, Read},
    net::TcpStream,
    str::{self},
};

const MAX_CONTENT_LENGTH: u64 = 1 << 10;

pub struct Request {
    pub http_version: String,
    pub method: HttpMethod,
    pub uri: String,
    pub headers: Vec<Header>,
    pub body: String,

    pub host: String,
    pub content_length: u64,

    pub parse_result: Result<()>,
}

pub enum HttpMethod {
    Get,
    Post,
    Put,
    Delete,
}

impl str::FromStr for HttpMethod {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "GET" => Ok(Self::Get),
            "POST" => Ok(Self::Post),
            "PUT" => Ok(Self::Put),
            "DELETE" => Ok(Self::Delete),
            _ => bail!("invalid method"),
        }
    }
}

impl fmt::Display for HttpMethod {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let method_str = match self {
            Self::Get => "GET",
            Self::Post => "POST",
            Self::Put => "PUT",
            Self::Delete => "Delete",
        };
        write!(f, "{method_str}")
    }
}

enum LineType {
    Start,
    Header,
    Body,
}

pub fn read_request(conn: &mut TcpStream) -> Result<Request> {
    let mut req = Request {
        http_version: "HTTP/1.1".to_string(),
        method: HttpMethod::Get,
        uri: String::new(),
        headers: vec![],
        body: String::new(),
        host: "".to_string(),
        content_length: 0,
        parse_result: Ok(()),
    };

    let mut br = BufReader::new(conn);
    let mut ltype = LineType::Start;
    for line in br.by_ref().lines() {
        let line = line.context("failed to read line")?;

        match ltype {
            LineType::Start => {
                let mut parts = line.splitn(3, ' ');
                let (Some(method), Some(uri), Some(http_version)) =
                    (parts.next(), parts.next(), parts.next())
                else {
                    req.parse_result = Err(anyhow!("malformed start line"));
                    return Ok(req);
                };
                let Ok(method) = method.parse::<HttpMethod>() else {
                    req.parse_result = Err(anyhow!("malformed start line"));
                    return Ok(req);
                };
                req.method = method;
                req.uri = uri.to_string();
                req.http_version = http_version.to_string();
                ltype = LineType::Header
            }
            LineType::Header => {
                #[allow(unused_assignments)]
                if line.is_empty() {
                    ltype = LineType::Body;
                    break; // header end
                }
                let mut parts = line.splitn(2, ':');
                let (Some(name), Some(value)) = (parts.next(), parts.next()) else {
                    req.parse_result = Err(anyhow!("malformed header"));
                    return Ok(req);
                };
                let h = Header {
                    name: name.trim().to_string(),
                    value: value.trim().trim_end_matches('/').to_string(),
                };
                req.headers.push(h.clone());
                match h.name.to_lowercase().as_str() {
                    "host" => req.host = h.value,
                    "content-length" => {
                        match h.value.parse() {
                            Ok(content_length) => req.content_length = content_length,
                            Err(e) => {
                                req.parse_result =
                                    Err(anyhow!("failed to parse Content-Length: {e}"));
                                return Ok(req);
                            }
                        };
                    }
                    _ => {}
                }
            }
            _ => unreachable!(),
        }
    }

    if req.content_length > MAX_CONTENT_LENGTH {
        req.parse_result = Err(anyhow!("Content-Length too large"));
        return Ok(req);
    }

    let mut buf = Vec::with_capacity(req.content_length as usize);
    br.take(req.content_length)
        .read_to_end(&mut buf)
        .context("failed to read body")?;
    req.body = String::from_utf8(buf).context("failed to convert body from utf8")?;

    Ok(req)
}

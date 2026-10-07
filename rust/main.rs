use anyhow::{Context, Result, anyhow, bail};
use std::{
    fmt::{self},
    io::{BufRead, BufReader, Read, Write},
    net::{TcpListener, TcpStream},
    str::{self, FromStr},
    thread::{self, JoinHandle},
};

const MAX_CONTENT_LENGTH: u64 = 1 << 10;

fn main() -> Result<()> {
    let addr = std::env::var("ADDRESS")
        .context("ADDRESS not set")?;

    let listener = TcpListener::bind(addr)?;

    let mut handles: Vec<JoinHandle<()>> = vec![];

    for conn in listener.incoming() {
        let mut conn = conn?;
        handles.push(thread::spawn(move || match handle_conn(&mut conn) {
            Ok(_) => (),
            Err(e) => eprintln!("failed to handle connection: {e}"),
        }));
    }

    for h in handles {
        h.join().map_err(|_| anyhow!("failed to join thread"))?;
    }

    Ok(())
}

fn handle_conn(conn: &mut TcpStream) -> Result<()> {
    let req = read_request(Read::by_ref(conn))?;
    let res = handle_request(&req);
    write_response(Write::by_ref(conn), &res)?;

    println!("{} {} {}", res.status_code, req.method, req.uri);

    Ok(())
}

struct Request {
    http_version: String,
    method: HttpMethod,
    uri: String,
    body: String,

    parse_result: Result<()>,
}

enum HttpMethod {
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

#[derive(Default)]
struct Response {
    http_version: String,
    status_code: u16,
    reason_pharse: String,
    body: String,
}

enum LineType {
    Start,
    Header,
    Body,
}

fn read_request(conn: &mut TcpStream) -> Result<Box<Request>> {
    let mut req = Box::new(Request {
        http_version: "HTTP/1.1".to_string(),
        method: HttpMethod::Get,
        uri: String::new(),
        body: String::new(),
        parse_result: Ok(()),
    });

    let mut br = BufReader::new(conn);
    let mut ltype = LineType::Start;
    for line in br.by_ref().lines() {
        let line = line.context("failed to read line")?;
        if line.is_empty() {
            break;
        }

        match ltype {
            LineType::Start => {
                let mut parts = line.splitn(3, ' ');
                let (Some(method), Some(uri), Some(http_version)) =
                    (parts.next(), parts.next(), parts.next())
                else {
                    req.parse_result = Err(anyhow!("malformed start line"));
                    return Ok(req);
                };
                match method.parse() {
                    Ok(method) => ,
                    Err(_) => ,
                }
                req.uri = uri.to_string();
                req.http_version = http_version.to_string();
                ltype = LineType::Header
            }
            LineType::Header => {}
            LineType::Body => todo!(),
        }
    }

    let content_len = MAX_CONTENT_LENGTH; // TODO: change to actual content length

    let mut buf = Vec::with_capacity(content_len as usize);
    br.take(content_len)
        .read_to_end(&mut buf)
        .context("failed to read body")?;
    req.body = String::from_utf8(buf).context("failed to convert body from utf8")?;

    Ok(req)
}

fn handle_request(req: &Request) -> Box<Response> {
    let mut res = Box::new(Response {
        http_version: "http/1.1".to_string(),
        status_code: 200,
        reason_pharse: "OK".to_string(),
        ..Default::default()
    });

    res.body = req.body.clone();

    res
}

fn write_response(conn: &mut TcpStream, res: &Response) -> Result<()> {
    // start line
    conn.write_fmt(format_args!(
        "{} {} {}",
        res.http_version, res.status_code, res.reason_pharse
    ))?;

    // body
    conn.write_fmt(format_args!("\r\n{}", res.body))?;

    conn.flush()?;

    Ok(())
}

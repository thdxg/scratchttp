use anyhow::Result;
use std::{io::Write, net::TcpStream};

use super::Header;

pub struct Response {
    pub http_version: &'static str,
    pub status_code: u16,
    pub reason_pharse: &'static str,
    pub headers: Vec<Header>,
    pub body: String,
}

pub fn write_response(conn: &mut TcpStream, res: &Response) -> Result<()> {
    // start line
    conn.write_fmt(format_args!(
        "{} {} {}\r\n",
        res.http_version, res.status_code, res.reason_pharse
    ))?;

    // headers
    for h in res.headers.iter() {
        conn.write_fmt(format_args!("{}: {}\r\n", h.name, h.value))?;
    }

    // body
    conn.write_fmt(format_args!("\r\n{}", res.body))?;

    conn.flush()?;

    Ok(())
}

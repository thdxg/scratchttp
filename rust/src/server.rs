pub mod db;
mod routes;
#[allow(clippy::module_inception)]
pub mod server;

pub use server::Server;

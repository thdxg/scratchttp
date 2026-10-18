use anyhow::{Result, bail};
use std::sync::{Arc, RwLock};

pub struct DB {
    data: Arc<RwLock<Vec<String>>>,
}

impl DB {
    pub fn new() -> Self {
        DB {
            data: Arc::new(RwLock::new(Vec::new())),
        }
    }

    pub fn add(&self, s: String) -> Result<()> {
        let Ok(mut data) = self.data.write() else {
            bail!("failed to acquire lock")
        };
        data.push(s);
        Ok(())
    }

    pub fn get(&self) -> Result<Vec<String>> {
        let Ok(data) = self.data.read() else {
            bail!("failed to acquire lock")
        };

        Ok(data.to_vec())
    }
}

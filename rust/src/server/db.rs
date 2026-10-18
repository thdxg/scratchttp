use anyhow::{Result, anyhow};
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
        let mut data = self
            .data
            .write()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;
        data.push(s);
        Ok(())
    }

    pub fn get(&self) -> Result<Vec<String>> {
        let data = self
            .data
            .read()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;

        Ok(data.to_vec())
    }
}

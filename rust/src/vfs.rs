use std::collections::HashMap;
use std::fs;
use std::io;
use std::sync::{Arc, RwLock};

pub struct VirtualFileSystem {
    files: RwLock<HashMap<String, Vec<u8>>>,
}

impl VirtualFileSystem {
    pub fn new() -> Self {
        Self {
            files: RwLock::new(HashMap::new()),
        }
    }

    pub fn write_file(&self, path: &str, content: Vec<u8>) {
        let mut files = self.files.write().unwrap();
        files.insert(path.to_string(), content);
    }

    pub fn read_file(&self, path: &str) -> io::Result<Vec<u8>> {
        let files = self.files.read().unwrap();
        if let Some(content) = files.get(path) {
            return Ok(content.clone());
        }

        fs::read(path)
    }

    pub fn flush(&self) -> io::Result<()> {
        let mut files = self.files.write().unwrap();
        for (path, content) in files.iter() {
            fs::write(path, content)?;
            println!("[hermes:vfs] Flushed: {}", path);
        }
        files.clear();
        Ok(())
    }
}

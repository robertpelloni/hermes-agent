use std::collections::HashMap;
use std::process::{Command, Child, Stdio};
use std::io::{self, Write, Read};
use std::sync::RwLock;

pub struct Plugin {
    pub name: String,
    process: Child,
}

pub struct PluginManager {
    plugins: RwLock<HashMap<String, Plugin>>,
}

impl PluginManager {
    pub fn new() -> Self {
        Self {
            plugins: RwLock::new(HashMap::new()),
        }
    }

    pub fn load(&self, name: &str, command: &str, args: &[&str]) -> io::Result<()> {
        let mut child = Command::new(command)
            .args(args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()?;

        let plugin = Plugin {
            name: name.to_string(),
            process: child,
        };

        let mut plugins = self.plugins.write().unwrap();
        plugins.insert(name.to_string(), plugin);
        println!("[hermes:mcp] Loaded plugin: {}", name);

        Ok(())
    }

    pub fn unload(&self, name: &str) -> io::Result<()> {
        let mut plugins = self.plugins.write().unwrap();
        if let Some(mut plugin) = plugins.remove(name) {
            plugin.process.kill()?;
            println!("[hermes:mcp] Unloaded plugin: {}", name);
        }
        Ok(())
    }
}

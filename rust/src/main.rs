use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// Start a chat session
    Chat {
        /// The message to send
        #[arg(short, long)]
        message: Option<String>,
    },
    /// Manage configuration
    Config {
        /// The key to set or get
        key: Option<String>,
    },
}

fn main() {
    let cli = Cli::parse();

    match &cli.command {
        Some(Commands::Chat { message }) => {
            if let Some(msg) = message {
                println!("Starting chat with message: {}", msg);
            } else {
                println!("Starting interactive chat session...");
            }
        }
        Some(Commands::Config { key }) => {
            if let Some(k) = key {
                println!("Reading config key: {}", k);
            } else {
                println!("Opening interactive config editor...");
            }
        }
        None => {
            println!("Hermes Agent CLI (Rust)");
            println!("Use --help for available commands.");
        }
    }
}

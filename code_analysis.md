# Code CLI Feature Analysis

The `just-every/code` repository is a fork of the Codex CLI that provides a command-line interface for the OpenAI API and other LLMs. Here are the key features and functionalities that need to be ported:

## Core Architecture
- **Language**: Rust (primarily `code-rs` and `codex-rs` directories) and some shell/bazel build scripts.
- **Build System**: Bazel with robust `MODULE.bazel` rules for Rust and Node.js.
- **Primary Interface**: CLI with subcommand routing (e.g., `code chat`, `code init`).

## Key Features to Port
1. **Model Interaction via CLI (`codex-cli`)**:
   - Sending prompts to language models.
   - Streaming responses to stdout with ANSI formatting.
   - Managing API keys via environment variables or configuration files.
2. **Context Management**:
   - Loading local files into the context window.
   - Reading repository state (git integration) to provide context.
3. **Session History**:
   - Storing past conversations locally.
   - Resuming previous sessions.
4. **Configuration System**:
   - `config.toml` support for defining default models, endpoints, and behaviors.

## Implementation Plan for 5 Languages

### Rust (`rust/`)
- Utilize `clap` for CLI parsing, similar to the original implementation.
- Use `tokio` for async HTTP requests to the model APIs.
- Re-implement the configuration parser using `serde`.

### Go (`go/`)
- Expand the existing Go launcher in `cmd/hermes` to support the full CLI surface of `code`.
- Use `cobra` for subcommand routing.
- Re-implement HTTP streaming using standard `net/http` and `bufio`.

### C# (`csharp/`)
- Use `System.CommandLine` for CLI parsing.
- Use `HttpClient` for async API requests.
- Implement configuration loading via `Microsoft.Extensions.Configuration`.

### Java (`java/`)
- Use `picocli` for command-line parsing.
- Use `java.net.http.HttpClient` for the API interactions.
- Implement a properties or JSON parser for the config.

### TypeScript (`typescript/`)
- Use `commander` or `yargs` for the CLI.
- Use `fetch` or `axios` for API calls.
- Share logic with the existing `ui-tui` if possible.

## Next Steps
- Write the foundational CLI parser in all 5 languages.
- Create dummy subcommands (`chat`, `config`) to match the `code` CLI interface.

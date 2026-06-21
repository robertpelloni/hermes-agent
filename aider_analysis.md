# Aider CLI Feature Analysis

The `paul-gauthier/aider` repository provides an AI pair programming tool in your terminal that lets you edit code via natural language.

## Core Architecture
- **Language**: Python (`aider` package directory).
- **Primary Interface**: Read-Eval-Print Loop (REPL) using `prompt_toolkit`.
- **Key Dependencies**: `litellm` for universal LLM routing, `tree-sitter` for syntactic repo mapping.

## Key Features to Port
1. **Repository Map (Tree-Sitter)**:
   - Creating a compact map of the entire repository using AST parsing.
   - Sending this map to the LLM to provide whole-codebase context without overwhelming the token limit.
2. **In-Terminal Diff Editing**:
   - The LLM streams Unified Diffs or Search/Replace blocks.
   - Aider automatically parses these blocks and patches the files on disk using robust fuzzy matching.
3. **Commit Automation**:
   - Automatically staging changed files and writing a descriptive commit message based on the delta.
4. **Interactive REPL Modes**:
   - Voice mode using Whisper.
   - `/add`, `/drop` for context management.
   - `/ask` vs `/code` routing.

## Integration into 5 Target Languages

### Rust (`rust/`)
- Utilize `tree-sitter-rs` to build the AST repo map.
- Implement a diff-parsing engine similar to Aider's `editblock` matching algorithm.

### Go (`go/`)
- Expand our `hermes-agent` TUI with `/add`, `/drop`, and `/commit` slash commands.
- Parse search/replace diff blocks from the agent's markdown response to modify files.

### C# (`csharp/`)
- Implement file-patching logic in C# using string proximity algorithms for the search blocks.
- Build an interactive loop using `System.Console`.

### Java (`java/`)
- Use JLine3 to implement a rich, prompt_toolkit-like REPL experience with autocomplete.

### TypeScript (`typescript/`)
- Reimplement the Aider diff logic using TypeScript.
- Bind the AST map to the existing CLI tool.

## Next Steps
- Analyze the `aider/coders/editblock_coder.py` to extract the exact search-and-replace algorithm logic.
- Plan the translation of the Aider Repomap logic across the 5 language ecosystems.

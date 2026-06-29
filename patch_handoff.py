import re

with open("HANDOFF.md", "r") as f:
    text = f.read()

text = re.sub(
    r"- \*\*Phase 2 Agent Loop\*\*: Implemented the native Go interactive REPL loop \(`internal/agent/loop.go`\) to handle full conversation flows. Enhanced the REPL to support `history` commands and basic context tracking logic.",
    r"- **Phase 2 Agent Loop**: Verified and documented the existing native Go interactive REPL loop (`cmd/hermes/main.go` and `pkg/agent/agent.go`) handles full conversation flows and stream processing.",
    text
)

text = re.sub(
    r"1\. Swap the `DummyMemory` in `internal/agent/loop\.go` with a full SQLite `GORM` persistence implementation\.",
    r"1. Finish verifying the robust auto-committing of LLM changes in Go.",
    text
)

with open("HANDOFF.md", "w") as f:
    f.write(text)

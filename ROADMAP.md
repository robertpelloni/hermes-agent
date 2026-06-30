# Project Roadmap

This roadmap outlines the major structural plans and strategic milestones for the Jules Autopilot Orchestrator.
For granular tasks and immediate bug fixes, see `TODO.md`.

## Milestone: v1.0.0 (Current) — "Deep Autonomous Node"
* [x] **Cross-Session Historical Intelligence**: The Autopilot now monitors for COMPLETED sessions, vectorizes the final result, and saves it into the `MemoryChunk` table for dual-layer RAG.
* [x] **Borg Discovery Handshake**: Added `GET /api/manifest` endpoint, broadcasting node capabilities and version for Borg assimilation.
* [x] **Session Replay Engine**: Added `GET /api/sessions/:id/replay` to provide a high-definition timeline of a session's entire history, optimized for Borg.
* [x] **Interactive Session Replay**: Integrated a `SessionReplayDialog` component accessible via a History icon on each session card.
* [x] **Global Fleet Heartbeat**: Added a "Fleet Pulse" section to the sidebar with a real-time active job counter and a pulsing brain icon.
* [x] **Autonomous Self-Healing**: The Autopilot actively monitors for the `FAILED` state, uses the Council Supervisor to analyze the error context, and autonomously messages Jules with a recovery plan.
* [x] **Visual Cognitive Status**: Session cards feature real-time "HEALING" and "EVALUATING" badges.
* [x] **Borg Fleet Summary API**: Implemented `GET /api/fleet/summary` for providing the Borg meta-orchestrator with a high-signal JSON payload of the fleet's state.
* [x] **Autonomous Issue Conversion**: Background daemon fetches open GitHub issues, evaluates if they are "Self-Healable", and autonomously spawns new Jules sessions.
* [x] **Continuous RAG Indexing**: Periodic background job chunks and embeds the repository into SQLite for "Long-Term Memory".
* [x] **Autonomous Multi-Agent Debates**: High-risk implementation plans trigger a background debate between a Security Architect and a Senior Engineer before auto-approval.
* [x] **Queue Telemetry**: Add deeper job-queue metrics.
* [x] **Git Diff Monitoring**: Background Shadow Pilot anomaly detection is missing native git diff monitoring.
* [x] **CI Pipeline Auto-Fix**: Shadow Pilot has anomaly logging but the CI pipeline auto-fix is incomplete.
* [x] **Submodule Status Check**: Real-time submodule git status checks in the Go backend are not fully wired to the `/system/status` UI.

## Phase 2: Agent Autonomy & Memory (Active)
- [x] Implement Agent Loop in Go (Interactive REPL loop).
- [x] Implement persistent Memory Storage structures in Go.
- [x] Implement robust auto-committing of LLM changes.
- [x] Implement Change Buffering Engine (In-Memory VFS).

## Parity Integrations
- [x] Implement AST Repository Mapping (Tree-sitter) in Rust, Go, C#, Java, TypeScript.
- [x] Implement Aider's Search/Replace Diff block patching engine.
- [x] Implement robust MCP (Model Context Protocol) plugin loading system across all languages.
- [x] Implement Dynamic Persona Configuration ("Fun Mode").

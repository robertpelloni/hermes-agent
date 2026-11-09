# Hermes Universal Configuration Architecture Specification

## 1. Overview & Architectural Principles

Hermes Agent is designed as a polyglot multi-harness system featuring a Python core engine, a Go native binary and subsystems layer, an Electron/TypeScript desktop environment, and a React web dashboard. To guarantee identical behavior, reproducible execution, and safe secret propagation across all runtime environments, Hermes adheres to a universal configuration contract.

### Core Precedence Hierarchy
Configuration values are resolved strictly from highest precedence to lowest:
1. **Explicit CLI Invocation Flags** (e.g., `--model`, `--provider`, `--accept-hooks`, `--port`)
2. **Runtime Environment Variables** (e.g., `HERMES_MODEL`, `HERMES_PROVIDER`, `HERMES_BASE_URL`)
3. **External Secret Vaults** (1Password `op://`, Bitwarden Secrets Manager `bws`)
4. **Local Environment File** (`~/.hermes/.env` or `.env` in working directory)
5. **Hierarchical YAML/JSON Configuration Files** (`~/.hermes/config.yaml`, project `./hermes.yaml`)
6. **Hardcoded Subsystem Defaults** (defined in core structs / dataclasses)

---

## 2. Universal Schema & Model Mapping

### 2.1 Model & Provider Routing
```yaml
# Primary model execution settings
model: "hermes-3-llama-3.1-70b"
provider: "nous"
base_url: "https://api.nousresearch.com/v1"
api_key: "${NOUS_API_KEY}"

# Auxiliary reasoning or fast models
auxiliary_model: "claude-3-5-haiku-latest"
auxiliary_provider: "anthropic"

# Fallback cascades for resilient failover
fallback_models:
  - model: "gpt-4o-mini"
    provider: "openai"
  - model: "qwen2.5-coder-32b-instruct"
    provider: "deepinfra"
```

### 2.2 Subsystem Configuration Matrix

| Canonical Config Key | Go Subsystem (`pkg/`) | Python Core (`hermes_cli/`) | Desktop / Web (`apps/desktop/`, `web/`) | Environment Override |
| :--- | :--- | :--- | :--- | :--- |
| `model` | `agent.Config.Model` | `config.active_model` | `useHermesConfig().model` | `HERMES_MODEL` |
| `provider` | `agent.Config.Provider` | `config.active_provider` | `useHermesConfig().provider` | `HERMES_PROVIDER` |
| `base_url` | `agent.Config.BaseURL` | `config.base_url` | `useHermesConfig().baseUrl` | `HERMES_BASE_URL` |
| `free_llm_port` | `agent.Config.BaseURL` fallback | `config.free_llm_port` | N/A | `HERMES_FREE_LLM_PORT` |
| `memory.db_path` | `memory.NewSQLiteStore(path)` | `memory.db` | Settings > Storage Path | `HERMES_MEMORY_DB` |
| `sessions.dir` | `sessionruntime.New(path)` | `sessions.storage_dir` | Session sidebar history | `HERMES_SESSIONS_DIR` |
| `skills.dirs` | `skill.LoadFromDir(path)` | `skills.search_paths` | Skill Hub catalog | `HERMES_SKILLS_PATH` |
| `mcp.servers` | `mcp.Server` / `adapters` | `mcp_config.servers` | Settings > MCP Servers | `HERMES_MCP_CONFIG` |
| `gateway.port` | `gateway.Gateway.Port` | `gateway.port` | Backend connection pool | `HERMES_GATEWAY_PORT` |
| `gateway.host` | `gateway.Gateway.Host` | `gateway.host` | Backend connection host | `HERMES_GATEWAY_HOST` |
| `compaction.token_limit` | `compaction.Compactor.Budget`| `compactor.max_tokens` | Chat context meter | `HERMES_TOKEN_LIMIT` |

---

## 3. External Secrets Management & Zero-Leakage Protocol

Hermes decouples runtime configuration from persistent credentials. Secrets are never checked into git or written unencrypted to disk.

### 3.1 1Password (`op CLI`) Integration
Resolves dynamic references formatted as `op://vault/item/field` during application bootstrap:
```yaml
secrets:
  onepassword:
    enabled: true
    account: "my-account"
    env:
      OPENAI_API_KEY: "op://Private/OpenAI/credential"
      ANTHROPIC_API_KEY: "op://Private/Anthropic/credential"
      NOUS_API_KEY: "op://Engineering/NousPortal/key"
```

### 3.2 Bitwarden Secrets Manager (`bws CLI`)
Bulk credential injection from central projects:
```yaml
secrets:
  bitwarden:
    enabled: true
    access_token_env: "BWS_ACCESS_TOKEN"
    project_id: "00000000-0000-0000-0000-000000000000"
    cache_ttl_seconds: 300
    override_existing: true
```

---

## 4. Cross-Harness Configuration Parity Validation

Every harness validates the parsed configuration schema against the canonical JSON/YAML definition at startup:
* **Go Subsystems:** Evaluated in `pkg/agent.DefaultConfig()` and `cmd/hermes/main.go`. Invalid model names or missing provider credentials fail fast before initiating network transports.
* **Python Engine:** Validated via Pydantic schema in `hermes_cli/config.py` with graceful fallback cascade.
* **Desktop App:** Validated via Zod schemas in `apps/desktop/src/app/session/hooks/use-hermes-config.ts`.

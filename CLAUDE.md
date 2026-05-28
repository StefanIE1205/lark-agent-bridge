# CLAUDE.md — lark-agent-bridge

## Project Overview

A local daemon service (Go) that bridges Lark (飞书) with local AI coding agents (Claude Code, Codex, Antigravity CLI). Users send natural language tasks from mobile Lark; the daemon receives messages via long connection, spawns/manages agent sessions in specified project directories, and reports progress/results/errors/diff summaries back to Lark.

**Module:** `github.com/StefanIE1205/lark-agent-bridge`
**Go version:** 1.23+
**Platform:** Windows-first (macOS/Linux later)
**Entry point:** `cmd/lab/main.go`

## Architecture — Module Boundaries (Non-Negotiable)

```
internal/lark/    → Lark long connection, message parsing, reply sending. NEVER depends on agent/.
internal/core/    → Command routing, session key generation, auth flow, task scheduling. Depends only on Platform interface, not Lark SDK.
internal/agent/   → Agent CLI start/input/output/stop. Knows nothing about Lark chat/thread/message.
internal/session/ → Session state machine, task status, active processes, recent logs.
internal/security/ → User authorization, high-risk action detection, approval flow, secret redaction.
internal/store/   → Atomic JSON file read/write for config state, project bindings, session metadata, audit log.
internal/config/  → TOML config loading and validation.
internal/logging/ → File-based logging setup.
```

**Rules:**
- `internal/lark` must NOT directly depend on `internal/agent` or any agent-specific code.
- `internal/core` must NOT directly call Lark SDK — it only uses the `Platform` interface.
- `internal/agent` must NOT know about Lark chat/thread/message structures.
- Each package must be independently testable with minimal mocking.

## Development Order (Strict)

Follow the task sequence exactly — each phase depends on the previous one:

1. **P0 (T001-T003):** Project skeleton — init, config loading, logging/data dirs
2. **P1 (T004-T006):** Lark message send/receive — long connection, parsing, `/ping` roundtrip
3. **P2 (T007-T010):** Core & commands — command parsing, auth, project/agent selection, session manager
4. **P3 (T011-T015):** Agent integration — fake agent → universal runner → Claude → Codex → Antigravity
5. **P4 (T016-T019):** Security & observability — redaction, approval, progress throttling, `/log` & `/sessions`
6. **P5 (T020-T022):** Release — Windows packaging, docs, build artifacts

**Never skip ahead.** Fake agent (T011) is the anchor for testing all downstream flows. Every agent after it is just a different CLI adapter.

## Code Conventions

### Go Style
- Follow standard Go idioms: small interfaces, explicit error returns, no panics in library code.
- Interface names: single-method interfaces use `-er` suffix (`Runner`, `Sender`); multi-method interfaces use descriptive names (`Platform`, `AgentSession`).
- File names: lowercase with underscores (`json_store.go`, `manager_test.go`).
- Package names: single word, lowercase, no underscores (`lark`, `core`, `agent`).
- One file = one primary concern. Split when a file exceeds ~400 lines.
- No init() functions except in `cmd/lab/main.go`.

### Error Handling
- All errors must be classified: user error / config error / agent error / platform error / internal error.
- User-facing errors: short and clear. Internal logs: full detail with context.
- Function signatures: `(T, error)` — no naked returns, no sentinel errors without wrapping context.
- Use `fmt.Errorf("...: %w", err)` for wrapping; use custom error types in `internal/core` only.

### Testing
- Test files alongside source: `foo.go` → `foo_test.go`.
- Must cover: config loading/validation, command parsing, Lark message text cleaning, session key generation, auth policy, redaction, atomic write.
- Integration tests use a fake agent: receives prompt, outputs N lines, exits. Long-running variant for stop tests.
- Manual acceptance: real Lark `/ping` roundtrip before moving past P1.

### Commands That Must Exist
- `go test ./...` — must always pass.
- `go run ./cmd/lab --version` — prints version.
- `go run ./cmd/lab --config config.toml` — starts the daemon.
- `go build -o lab.exe ./cmd/lab` — produces single binary.

## Security — Always Active

- All external input is untrusted. Validate at boundaries.
- Local command execution is **off by default**; user must explicitly enable via config.
- Non-admin users cannot execute `/bind`, `/ask`, `/stop`, `/approve`, `/deny`.
- Group chats not in `allowed_chat_ids` are silently ignored.
- Group chats without `@bot` mention are ignored (unless config overrides).
- `message_id` deduplication is mandatory.
- Secrets in logs/replies must be redacted: `API_KEY`, `TOKEN`, `SECRET`, `PASSWORD`, `CREDENTIAL`, `.env` values.
- Sanitize the `Redact()` function output: replace values with `***`.

## MVP Scope — What NOT to Build

- No SaaS/multi-tenant.
- No platforms other than Lark.
- No full web admin UI.
- No plugin marketplace.
- No complex Lark card UI — text messages with `/approve <id>` and `/deny <id>`.
- No database — JSON files with atomic writes only.
- Antigravity is experimental; don't let it block Claude/Codex progress.
- Session log rotation: MVP can skip, but `/log` must cap at 200 lines.
- PTY: Windows is complex; use stdin/stdout pipes for MVP.

## Configuration

Path lookup order: `--config <path>` → `./config.toml` → `~/.lark-agent-bridge/config.toml`

Required: `lark.app_id`, `lark.app_secret`, at least one admin user.
Validation on startup: config is legal, data_dir is writable, at least one agent command is executable, project paths exist.

## Key Dependencies

- `github.com/larksuite/oapi-sdk-go/v3` — Lark official SDK (WebSocket long connection)
- `github.com/BurntSushi/toml` — TOML config parsing
- `golang.org/x/sync` — errgroup for concurrent session management
- Standard library: `os/exec`, `context`, `encoding/json`, `sync`

## Git Workflow

- Branch naming: `task/TXXX-short-description` (e.g., `task/T001-init-repo`).
- Commit messages: conventional commits (`feat:`, `fix:`, `chore:`, `test:`).
- No force push to main.
- PR must include test run output.

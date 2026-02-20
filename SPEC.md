# clawtop — spec (v0)

## Scope (v0)

One-screen TUI similar to `htop`:

- Status of OpenClaw sessions/subagents (with basic filters)
- Task history (log-level filter)
- Next crons + last run status
- Token history (hourly/daily)
- Host CPU/memory/disk load
- Claude Code status (running? totals from `~/.claude.json`)

### Non-goals (v0)

- No actions / control plane (read-only)
- No credentials in repo
- GPU metrics optional, later

## Data sources (auto-discovery)

Defaults:
- `~/.openclaw/agents/main/sessions/sessions.json`
- `~/.openclaw/cron/jobs.json`
- `~/.openclaw/cron/runs/*.jsonl`
- `~/.openclaw/subagents/runs.json`
- `~/.openclaw/workspace/dashboard/metrics/tokens.jsonl` (optional)
- `~/.claude.json` (optional)

Overrides:
- `--openclaw-root <path>`
- `--workspace <path>`
- `--claude-config <path>`

## UX

- Default refresh: 2s
- Keys: `q` quit, `r` refresh now, `+/-` adjust refresh
- Filters: error/warn/info/debug

## Packaging

- MIT License
- Public repo `cl4wb0rg/clawtop`
- GitHub Actions releases: linux/amd64 + linux/arm64

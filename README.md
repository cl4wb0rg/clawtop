# clawtop

A minimal, read-only TUI (htop-style) for OpenClaw runtime status.

## Install

```bash
go install github.com/cl4wb0rg/clawtop/cmd/clawtop@latest
```

Or build from source:

```bash
go build -o clawtop ./cmd/clawtop
```

## Usage

```bash
clawtop
```

Flags:

- `--openclaw-root <path>` (default: `~/.openclaw` or `$OPENCLAW_ROOT`)
- `--workspace <path>` (default: `<openclaw-root>/workspace`)
- `--refresh 2s`

## Keys

- `q` / `Ctrl+C` quit
- `r` refresh now
- `+` / `-` faster / slower refresh

Toggles:

- `1` current reality (only sessions updated in last 24h)
- `2` hide `:run:` sessions
- `3` primary model only

Task filters:

- Levels: `e` error, `w` warn, `i` info, `d` debug
- Sources: `c` cron, `s` subagent, `t` tool

## Data sources (auto-discovery)

Defaults:

- `~/.openclaw/agents/main/sessions/sessions.json`
- `~/.openclaw/subagents/runs.json`
- `~/.openclaw/cron/jobs.json`
- `~/.openclaw/cron/runs/*.jsonl`
- `<workspace>/dashboard/metrics/tokens.jsonl` (optional)

## Status

MVP: single-screen dashboard with sessions/subagents, tasks, crons, tokens, host CPU/mem/load.

Read-only by design.

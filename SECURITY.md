# SECURITY.md — clawtop

## Baseline principles (inlined)

- **Least privilege**: no unnecessary permissions, no root.
- **No secrets in repo**: no tokens/keys/passwords in code, docs, or systemd units.
- **Safe by default**: explicit opt-in for risky flags/features.
- **Defensive parsing**: inputs treated as untrusted; set limits, handle errors cleanly.
- **No `eval` / unsafe shelling-out**: subprocesses use argument lists only.
- **Logging**: minimize sensitive data in logs; debug logs only when needed.
- **Dependency hygiene**: keep deps minimal; run vuln scans periodically.

---

## Repo-specific notes

**Type:** Terminal UI tool (TUI) — read-only system monitoring, no network services,
no persistent storage, no credentials.

**Threat model:**
- No secrets handled; no tokens, keys, or passwords are used or stored.
- Input comes only from local system APIs (procfs, sysfs, OS metrics) and CLI flags —
  no external/untrusted network input.
- No shell-out with user-controlled strings; all subprocess calls (if any) use
  argument lists.

**Dependencies:** Go standard library + Bubble Tea / Lip Gloss; kept minimal.
Run `govulncheck ./...` to check for known vulnerabilities.

**Binary distribution:** Statically compiled; no installer scripts.

---

## Vulnerability reporting

Report security issues **privately** — do not open a public issue for vulnerabilities.
Include repro steps, impact, and affected versions.

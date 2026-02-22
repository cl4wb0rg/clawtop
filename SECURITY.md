# SECURITY.md — clawtop

This project follows the workspace-wide security baseline defined in
[`../SECURITY_BASELINE.md`](../SECURITY_BASELINE.md).

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

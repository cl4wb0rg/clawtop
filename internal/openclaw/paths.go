package openclaw

import (
	"errors"
	"os"
	"path/filepath"
)

// Paths are resolved locations for OpenClaw state.
// All paths are local and read-only for clawtop.
//
// Most users will rely on auto-discovery:
//   OpenClawRoot: ~/.openclaw
//   WorkspaceDir: <OpenClawRoot>/workspace
//
// Overrides are supported via CLI flags.
type Paths struct {
	OpenClawRoot string
	WorkspaceDir string

	SessionsJSON string
	SubagentRuns string
	CronJobs     string
	CronRunsDir  string
	TokensJSONL  string
}

// DiscoverPaths resolves OpenClawRoot and related file locations.
//
// If openclawRootOverride is empty, it defaults to ~/.openclaw (or $OPENCLAW_ROOT).
// If workspaceOverride is empty, it defaults to <openclawRoot>/workspace.
//
// The returned paths may point to non-existent optional files (e.g. tokens.jsonl).
func DiscoverPaths(openclawRootOverride, workspaceOverride string) (Paths, error) {
	root := openclawRootOverride
	if root == "" {
		root = os.Getenv("OPENCLAW_ROOT")
	}
	if root == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		root = filepath.Join(h, ".openclaw")
	}
	root, _ = filepath.Abs(root)

	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		return Paths{}, errors.New("openclaw root not found: " + root)
	}

	ws := workspaceOverride
	if ws == "" {
		ws = filepath.Join(root, "workspace")
	}
	ws, _ = filepath.Abs(ws)

	p := Paths{
		OpenClawRoot: root,
		WorkspaceDir: ws,
		SessionsJSON: filepath.Join(root, "agents", "main", "sessions", "sessions.json"),
		SubagentRuns: filepath.Join(root, "subagents", "runs.json"),
		CronJobs:     filepath.Join(root, "cron", "jobs.json"),
		CronRunsDir:  filepath.Join(root, "cron", "runs"),
		TokensJSONL:  filepath.Join(ws, "dashboard", "metrics", "tokens.jsonl"),
	}
	return p, nil
}

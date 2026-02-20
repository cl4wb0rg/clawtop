package openclaw

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPaths_DefaultsToHomeOpenClaw(t *testing.T) {
	home := t.TempDir()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, ".openclaw")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := DiscoverPaths("", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.OpenClawRoot != root {
		t.Fatalf("OpenClawRoot=%q want %q", p.OpenClawRoot, root)
	}
	if p.SessionsJSON == "" || p.CronJobs == "" || p.SubagentRuns == "" {
		t.Fatalf("expected core paths to be set: %#v", p)
	}
}

func TestDiscoverPaths_Override(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(root, "ws")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	p, err := DiscoverPaths(root, ws)
	if err != nil {
		t.Fatal(err)
	}
	if p.OpenClawRoot != root {
		t.Fatalf("OpenClawRoot=%q want %q", p.OpenClawRoot, root)
	}
	if p.WorkspaceDir != ws {
		t.Fatalf("WorkspaceDir=%q want %q", p.WorkspaceDir, ws)
	}
}

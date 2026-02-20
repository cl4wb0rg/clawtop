package openclaw

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadSessionsJSON(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "sessions.json")
	b := []byte(`{
  "agent:main:main": {"label":"Main","model":"gpt-5.2","modelProvider":"openai-codex","updatedAt": 1700000000000, "inputTokens": 1, "outputTokens": 2, "totalTokens": 3},
  "agent:main:cron:1": {"label":"Cron","model":"gpt-5.2","modelProvider":"openai-codex","updatedAt": 1600000000000}
}`)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := ReadSessionsJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 2 {
		t.Fatalf("len=%d", len(s))
	}
	if s[0].Key != "agent:main:main" {
		t.Fatalf("sorted by updatedAt desc, got first=%s", s[0].Key)
	}
}

func TestReadCronJobs(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "jobs.json")
	b := []byte(`{
  "version": 1,
  "jobs": [
    {"id":"a","name":"job-a","enabled":true,
     "schedule":{"kind":"cron","expr":"* * * * *","tz":"UTC"},
     "state":{"nextRunAtMs": 1700000000000, "lastRunAtMs": 1690000000000, "lastStatus":"ok"}}
  ]
}`)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	jobs, err := ReadCronJobs(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("len=%d", len(jobs))
	}
	if jobs[0].NextRun == nil || jobs[0].NextRun.UnixMilli() != 1700000000000 {
		t.Fatalf("NextRun=%v", jobs[0].NextRun)
	}
}

func TestReadTokenSamples(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "tokens.jsonl")
	b := []byte("" +
		`{"ts":1700000000000,"openclaw":{"total":10},"claudeCode":{"costUSD":1.2}}` + "\n" +
		`{"ts":1700003600000,"openclaw":{"total":20},"claudeCode":{"costUSD":2.2}}` + "\n")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := ReadTokenSamples(p, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 2 {
		t.Fatalf("len=%d", len(s))
	}
	if !s[0].At.Before(s[1].At) {
		t.Fatal("expected chronological")
	}
}

func TestReadToolTasks(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "sess.jsonl")
	b := []byte("" +
		`{"type":"message","message":{"role":"toolResult","toolName":"exec","isError":false,"timestamp":1700000000000,"content":[{"type":"text","text":"ok\nmore"}]}}` + "\n" +
		`{"type":"message","message":{"role":"toolResult","toolName":"write","isError":true,"timestamp":1700000001000,"content":[{"type":"text","text":"boom"}]}}` + "\n")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	tasks, err := ReadToolTasks(p, 50, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len=%d", len(tasks))
	}
	if tasks[1].Level != LevelError {
		t.Fatalf("level=%s", tasks[1].Level)
	}
	if tasks[0].At.After(tasks[1].At) {
		t.Fatal("expected chronological")
	}
	if tasks[0].Source != SourceTool {
		t.Fatalf("source=%s", tasks[0].Source)
	}
	_ = time.Second
}

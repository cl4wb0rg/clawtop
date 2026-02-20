package openclaw

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func ReadSessionsJSON(path string) ([]Session, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// sessions.json is a map[sessionKey]sessionState
	var raw map[string]struct {
		Label        string `json:"label"`
		Model        string `json:"model"`
		ModelProvider string `json:"modelProvider"`
		UpdatedAt    int64  `json:"updatedAt"`
		InputTokens  int64  `json:"inputTokens"`
		OutputTokens int64  `json:"outputTokens"`
		TotalTokens  int64  `json:"totalTokens"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]Session, 0, len(raw))
	for k, v := range raw {
		out = append(out, Session{
			Key:          k,
			Label:        v.Label,
			Model:        v.Model,
			Provider:     v.ModelProvider,
			UpdatedAt:    time.UnixMilli(v.UpdatedAt),
			InputTokens:  v.InputTokens,
			OutputTokens: v.OutputTokens,
			TotalTokens:  v.TotalTokens,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func ReadSubagentRuns(path string) ([]SubagentRun, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Version int `json:"version"`
		Runs    map[string]struct {
			RunID           string `json:"runId"`
			ChildSessionKey string `json:"childSessionKey"`
			Label           string `json:"label"`
			Task            string `json:"task"`
			Model           string `json:"model"`
			CreatedAt       int64  `json:"createdAt"`
			StartedAt       *int64 `json:"startedAt"`
			FinishedAt      *int64 `json:"finishedAt"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]SubagentRun, 0, len(raw.Runs))
	for _, r := range raw.Runs {
		s := SubagentRun{
			RunID:           r.RunID,
			ChildSessionKey: r.ChildSessionKey,
			Label:           r.Label,
			Task:            r.Task,
			Model:           r.Model,
			CreatedAt:       time.UnixMilli(r.CreatedAt),
		}
		if r.StartedAt != nil {
			t := time.UnixMilli(*r.StartedAt)
			s.StartedAt = &t
		}
		if r.FinishedAt != nil {
			t := time.UnixMilli(*r.FinishedAt)
			s.FinishedAt = &t
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func ReadCronJobs(path string) ([]CronJob, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Version int `json:"version"`
		Jobs    []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
			Schedule struct {
				Kind string `json:"kind"`
				Expr string `json:"expr"`
				TZ   string `json:"tz"`
				At   string `json:"at"`
				EveryMs int64 `json:"everyMs"`
			} `json:"schedule"`
			State struct {
				NextRunAtMs int64  `json:"nextRunAtMs"`
				LastRunAtMs int64  `json:"lastRunAtMs"`
				LastStatus  string `json:"lastStatus"`
				LastError   string `json:"lastError"`
			} `json:"state"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]CronJob, 0, len(raw.Jobs))
	for _, j := range raw.Jobs {
		cj := CronJob{ID: j.ID, Name: j.Name, Enabled: j.Enabled, TZ: j.Schedule.TZ, LastStatus: j.State.LastStatus, LastError: j.State.LastError}
		sched := j.Schedule.Kind
		switch j.Schedule.Kind {
		case "cron":
			sched = j.Schedule.Expr
		case "at":
			sched = "at " + j.Schedule.At
		case "every":
			sched = fmt.Sprintf("every %s", (time.Duration(j.Schedule.EveryMs)*time.Millisecond).String())
		}
		cj.Schedule = sched
		if j.State.NextRunAtMs != 0 {
			t := time.UnixMilli(j.State.NextRunAtMs)
			cj.NextRun = &t
		}
		if j.State.LastRunAtMs != 0 {
			t := time.UnixMilli(j.State.LastRunAtMs)
			cj.LastRun = &t
		}
		out = append(out, cj)
	}
	sort.Slice(out, func(i, j int) bool {
		// Next run ascending, nil last
		if out[i].NextRun == nil {
			return false
		}
		if out[j].NextRun == nil {
			return true
		}
		return out[i].NextRun.Before(*out[j].NextRun)
	})
	return out, nil
}

// ReadLatestCronRun reads the most recent "finished" event from a cron runs jsonl file.
// It returns status and either error or first line of summary.
func ReadLatestCronRun(path string) (Task, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return Task{}, false, err
	}
	defer f.Close()

	lines, err := tailLines(f, 200)
	if err != nil {
		return Task{}, false, err
	}
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var rec struct {
			TS     int64  `json:"ts"`
			Action string `json:"action"`
			Status string `json:"status"`
			Error  string `json:"error"`
			Summary string `json:"summary"`
			JobID  string `json:"jobId"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		if rec.Action != "finished" {
			continue
		}
		lvl := LevelInfo
		if rec.Status == "error" {
			lvl = LevelError
		}
		detail := firstLine(rec.Summary)
		if rec.Error != "" {
			detail = rec.Error
		}
		return Task{At: time.UnixMilli(rec.TS), Level: lvl, Source: SourceCron, Title: rec.JobID, Detail: detail}, true, nil
	}
	return Task{}, false, nil
}

func ReadTokenSamples(path string, max int) ([]TokenSample, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	samples := make([]TokenSample, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		var rec struct {
			TS int64 `json:"ts"`
			OpenClaw struct {
				Total int64 `json:"total"`
			} `json:"openclaw"`
			ClaudeCode struct {
				CostUSD float64 `json:"costUSD"`
			} `json:"claudeCode"`
		}
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			continue
		}
		samples = append(samples, TokenSample{At: time.UnixMilli(rec.TS), OpenClawTotal: rec.OpenClaw.Total, ClaudeCostUSD: rec.ClaudeCode.CostUSD})
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].At.Before(samples[j].At) })
	if max > 0 && len(samples) > max {
		samples = samples[len(samples)-max:]
	}
	return samples, nil
}

// ReadToolTasks reads recent tool results from a session jsonl file.
func ReadToolTasks(sessionJSONL string, maxLines int, maxTasks int) ([]Task, error) {
	f, err := os.Open(sessionJSONL)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lines, err := tailLines(f, maxLines)
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, 64)
	for i := len(lines) - 1; i >= 0 && len(tasks) < maxTasks; i-- {
		ln := strings.TrimSpace(lines[i])
		if ln == "" {
			continue
		}
		var rec struct {
			Type    string `json:"type"`
			Message struct {
				Role string `json:"role"`
				ToolName string `json:"toolName"`
				IsError bool `json:"isError"`
				Timestamp int64 `json:"timestamp"`
				Content []struct{ Type string `json:"type"`; Text string `json:"text"` } `json:"content"`
			} `json:"message"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			continue
		}
		if rec.Type != "message" || rec.Message.Role != "toolResult" {
			continue
		}
		lvl := LevelInfo
		if rec.Message.IsError {
			lvl = LevelError
		}
		at := time.Now()
		if rec.Message.Timestamp != 0 {
			at = time.UnixMilli(rec.Message.Timestamp)
		}
		detail := ""
		if len(rec.Message.Content) > 0 {
			detail = firstLine(rec.Message.Content[0].Text)
		}
		tasks = append(tasks, Task{At: at, Level: lvl, Source: SourceTool, Title: rec.Message.ToolName, Detail: detail})
	}
	// reverse to chronological
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
	return tasks, nil
}

func CronRunFile(cronRunsDir, jobID string) string {
	return filepath.Join(cronRunsDir, jobID+".jsonl")
}

func firstLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}

// tailLines reads up to n last lines from r (which must be an *os.File).
func tailLines(f *os.File, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	// Read up to last 256KB; good enough for our state files.
	const max = 256 * 1024
	start := int64(0)
	if st.Size() > max {
		start = st.Size() - max
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}
	r := bufio.NewReader(f)
	if start > 0 {
		// discard partial line
		_, _ = r.ReadString('\n')
	}
	lines := make([]string, 0, 1024)
	for {
		ln, err := r.ReadString('\n')
		if errors.Is(err, io.EOF) {
			if strings.TrimSpace(ln) != "" {
				lines = append(lines, strings.TrimRight(ln, "\n"))
			}
			break
		}
		if err != nil {
			return nil, err
		}
		lines = append(lines, strings.TrimRight(ln, "\n"))
	}
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}

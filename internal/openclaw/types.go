package openclaw

import "time"

type Session struct {
	Key         string
	Label       string
	Model       string
	Provider    string
	UpdatedAt   time.Time
	InputTokens int64
	OutputTokens int64
	TotalTokens int64
}

type SubagentRun struct {
	RunID           string
	ChildSessionKey string
	Label           string
	Task            string
	Model           string
	CreatedAt       time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
}

type CronJob struct {
	ID        string
	Name      string
	Enabled   bool
	Schedule  string
	TZ        string
	NextRun   *time.Time
	LastRun   *time.Time
	LastStatus string
	LastError  string
}

type TaskLevel string

const (
	LevelError TaskLevel = "error"
	LevelWarn  TaskLevel = "warn"
	LevelInfo  TaskLevel = "info"
	LevelDebug TaskLevel = "debug"
)

type TaskSource string

const (
	SourceCron    TaskSource = "cron"
	SourceSubagent TaskSource = "subagent"
	SourceTool    TaskSource = "tool"
)

type Task struct {
	At      time.Time
	Level   TaskLevel
	Source  TaskSource
	Title   string
	Detail  string
}

type TokenSample struct {
	At time.Time
	OpenClawTotal int64
	ClaudeCostUSD float64
}

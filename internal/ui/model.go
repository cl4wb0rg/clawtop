package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cl4wb0rg/clawtop/internal/host"
	"github.com/cl4wb0rg/clawtop/internal/openclaw"
)

type Config struct {
	Paths openclaw.Paths
	Refresh time.Duration
}

type model struct {
	cfg Config

	width int
	height int

	refresh time.Duration
	lastUpdate time.Time
	err error

	// host stats
	prevCPU *host.CPUStat
	host host.HostMetrics

	// openclaw data
	sessions []openclaw.Session
	subagents []openclaw.SubagentRun
	crons []openclaw.CronJob
	tasks []openclaw.Task
	tokenSamples []openclaw.TokenSample

	// filters/toggles
	filter24h bool
	hideRunSessions bool
	primaryModelOnly bool
	primaryModel string

	levels map[openclaw.TaskLevel]bool
	sources map[openclaw.TaskSource]bool
}

type tickMsg time.Time

type refreshMsg struct {
	err error
	at time.Time
	sessions []openclaw.Session
	subagents []openclaw.SubagentRun
	crons []openclaw.CronJob
	tasks []openclaw.Task
	tokenSamples []openclaw.TokenSample
	host host.HostMetrics
	cpu host.CPUStat
	hasCPU bool
}

func New(cfg Config) tea.Model {
	m := model{cfg: cfg, refresh: cfg.Refresh}
	if m.refresh <= 0 {
		m.refresh = 2 * time.Second
	}
	m.levels = map[openclaw.TaskLevel]bool{openclaw.LevelError: true, openclaw.LevelWarn: true, openclaw.LevelInfo: true, openclaw.LevelDebug: false}
	m.sources = map[openclaw.TaskSource]bool{openclaw.SourceCron: true, openclaw.SourceSubagent: true, openclaw.SourceTool: true}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshNowCmd(),
		tickCmd(m.refresh),
	)
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.refreshNowCmd(), tickCmd(m.refresh))
	case refreshMsg:
		m.lastUpdate = msg.at
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.sessions = msg.sessions
		m.subagents = msg.subagents
		m.crons = msg.crons
		m.tasks = msg.tasks
		m.tokenSamples = msg.tokenSamples
		m.host = msg.host
		if msg.hasCPU {
			m.prevCPU = &msg.cpu
		}
		if m.primaryModel == "" {
			m.primaryModel = guessPrimaryModel(m.sessions)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, m.refreshNowCmd()
		case "+":
			if m.refresh > 500*time.Millisecond {
				m.refresh -= 500 * time.Millisecond
			}
			return m, nil
		case "-":
			m.refresh += 500 * time.Millisecond
			return m, nil
		case "1":
			m.filter24h = !m.filter24h
			return m, nil
		case "2":
			m.hideRunSessions = !m.hideRunSessions
			return m, nil
		case "3":
			m.primaryModelOnly = !m.primaryModelOnly
			return m, nil
		case "e":
			m.levels[openclaw.LevelError] = !m.levels[openclaw.LevelError]
			return m, nil
		case "w":
			m.levels[openclaw.LevelWarn] = !m.levels[openclaw.LevelWarn]
			return m, nil
		case "i":
			m.levels[openclaw.LevelInfo] = !m.levels[openclaw.LevelInfo]
			return m, nil
		case "d":
			m.levels[openclaw.LevelDebug] = !m.levels[openclaw.LevelDebug]
			return m, nil
		case "c":
			m.sources[openclaw.SourceCron] = !m.sources[openclaw.SourceCron]
			return m, nil
		case "s":
			m.sources[openclaw.SourceSubagent] = !m.sources[openclaw.SourceSubagent]
			return m, nil
		case "t":
			m.sources[openclaw.SourceTool] = !m.sources[openclaw.SourceTool]
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("clawtop")
	sub := fmt.Sprintf("refresh=%s  updated=%s", m.refresh, relTime(m.lastUpdate))
	if m.err != nil {
		sub += "  err=" + m.err.Error()
	}

	filters := fmt.Sprintf("[1]24h:%s [2]hide:run:%s [3]primary(%s):%s  levels e:%s w:%s i:%s d:%s  src c:%s s:%s t:%s  r refresh  +/- rate  q quit",
		onOff(m.filter24h), onOff(m.hideRunSessions), m.primaryModel, onOff(m.primaryModelOnly),
		onOff(m.levels[openclaw.LevelError]), onOff(m.levels[openclaw.LevelWarn]), onOff(m.levels[openclaw.LevelInfo]), onOff(m.levels[openclaw.LevelDebug]),
		onOff(m.sources[openclaw.SourceCron]), onOff(m.sources[openclaw.SourceSubagent]), onOff(m.sources[openclaw.SourceTool]),
	)

	leftW := m.width/2 - 1
	if leftW < 40 {
		leftW = 40
	}
	left := lipgloss.NewStyle().Width(leftW)
	right := lipgloss.NewStyle().Width(m.width - leftW - 1)

	leftBody := strings.Join([]string{
		renderHost(m.host),
		renderTokens(m.tokenSamples),
		renderSessions(m.sessions, m.subagents, sessionFilters{only24h: m.filter24h, hideRun: m.hideRunSessions, primaryModelOnly: m.primaryModelOnly, primaryModel: m.primaryModel}),
	}, "\n\n")

	rightBody := strings.Join([]string{
		renderTasks(m.tasks, taskFilters{levels: m.levels, sources: m.sources}),
		renderCrons(m.crons),
	}, "\n\n")

	body := lipgloss.JoinHorizontal(lipgloss.Top, left.Render(leftBody), right.Render(rightBody))

	return strings.Join([]string{header + "  " + sub, filters, body}, "\n") + "\n"
}

func (m model) refreshNowCmd() tea.Cmd {
	paths := m.cfg.Paths
	prevCPU := m.prevCPU
	return func() tea.Msg {
		at := time.Now()
		var out refreshMsg
		out.at = at

		// host
		cpu, err := host.ReadCPUStat()
		if err == nil {
			out.cpu = cpu
			out.hasCPU = true
		}
		l1, l5, l15, _ := host.ReadLoadAvg()
		total, avail, _ := host.ReadMemInfo()
		used := total - avail
		cpuPct := 0.0
		if prevCPU != nil && out.hasCPU {
			cpuPct = host.CPUPercent(*prevCPU, cpu)
		}
		out.host = host.HostMetrics{At: at, CPUPercent: cpuPct, MemUsedBytes: used, MemTotalBytes: total, Load1: l1, Load5: l5, Load15: l15}

		// openclaw
		if sessions, err := openclaw.ReadSessionsJSON(paths.SessionsJSON); err == nil {
			out.sessions = sessions
		} else {
			out.err = err
			return out
		}
		if sub, err := openclaw.ReadSubagentRuns(paths.SubagentRuns); err == nil {
			out.subagents = sub
		}
		if cr, err := openclaw.ReadCronJobs(paths.CronJobs); err == nil {
			out.crons = cr
		}

		// tasks: cron finished + tool results + subagent runs
		tasks := make([]openclaw.Task, 0, 64)
		if st, err := os.Stat(paths.CronRunsDir); err == nil && st.IsDir() {
			for _, cj := range out.crons {
				p := openclaw.CronRunFile(paths.CronRunsDir, cj.ID)
				if _, err := os.Stat(p); err == nil {
					if t, ok, _ := openclaw.ReadLatestCronRun(p); ok {
						t.Title = "cron: " + cj.Name
						tasks = append(tasks, t)
					}
				}
			}
		}
		// tool tasks from main session file if present in sessions.json
		mainSess := findSession(out.sessions, "agent:main:main")
		if mainSess != nil {
			// session file is in sessions.json as sessionFile but we don't parse it. fallback to default.
			// for MVP we assume default location.
			sessionFile := paths.OpenClawRoot + "/agents/main/sessions/" + "" // unused
			_ = sessionFile
		}
		// heuristic: pick newest agent:main:main session file by glob
		glob := paths.OpenClawRoot + "/agents/main/sessions/*.jsonl"
		matches, _ := filepathGlob(glob)
		sort.Slice(matches, func(i, j int) bool {
			iSt, _ := os.Stat(matches[i])
			jSt, _ := os.Stat(matches[j])
			if iSt == nil || jSt == nil {
				return matches[i] > matches[j]
			}
			return iSt.ModTime().After(jSt.ModTime())
		})
		if len(matches) > 0 {
			if toolTasks, err := openclaw.ReadToolTasks(matches[0], 400, 25); err == nil {
				tasks = append(tasks, toolTasks...)
			}
		}
		for _, sa := range out.subagents {
			lvl := openclaw.LevelInfo
			detail := firstN(sa.Task, 90)
			tasks = append(tasks, openclaw.Task{At: sa.CreatedAt, Level: lvl, Source: openclaw.SourceSubagent, Title: "subagent: " + sa.Label, Detail: detail})
		}
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].At.After(tasks[j].At) })
		if len(tasks) > 40 {
			tasks = tasks[:40]
		}
		out.tasks = tasks

		// tokens optional
		if samples, err := openclaw.ReadTokenSamples(paths.TokensJSONL, 48); err == nil {
			out.tokenSamples = samples
		}

		return out
	}
}

// filepathGlob wrapped for testability.
var filepathGlob = func(pattern string) ([]string, error) { return filepath.Glob(pattern) }

func findSession(s []openclaw.Session, key string) *openclaw.Session {
	for i := range s {
		if s[i].Key == key {
			return &s[i]
		}
	}
	return nil
}

func relTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t)
	if d < time.Second {
		return "now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func guessPrimaryModel(s []openclaw.Session) string {
	for _, sess := range s {
		if sess.Key == "agent:main:main" && sess.Model != "" {
			return sess.Model
		}
	}
	// fallback most common
	m := map[string]int{}
	for _, sess := range s {
		if sess.Model != "" {
			m[sess.Model]++
		}
	}
	best := ""
	bestN := 0
	for k, n := range m {
		if n > bestN {
			best, bestN = k, n
		}
	}
	return best
}

func firstN(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

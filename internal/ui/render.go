package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/cl4wb0rg/clawtop/internal/host"
	"github.com/cl4wb0rg/clawtop/internal/openclaw"
)

type sessionFilters struct {
	only24h bool
	hideRun bool
	primaryModelOnly bool
	primaryModel string
}

type taskFilters struct {
	levels map[openclaw.TaskLevel]bool
	sources map[openclaw.TaskSource]bool
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
)

func renderHost(m host.HostMetrics) string {
	return titleStyle.Render("Host") + "\n" +
		fmt.Sprintf("CPU: %5.1f%%   Mem: %s/%s   Load: %.2f %.2f %.2f",
			m.CPUPercent,
			host.HumanBytes(m.MemUsedBytes), host.HumanBytes(m.MemTotalBytes),
			m.Load1, m.Load5, m.Load15,
		)
}

func renderTokens(samples []openclaw.TokenSample) string {
	if len(samples) == 0 {
		return titleStyle.Render("Tokens") + "\n" + dimStyle.Render("(no tokens.jsonl)")
	}
	last := samples[len(samples)-1]
	spark := sparkline(samples)
	return titleStyle.Render("Tokens") + "\n" +
		fmt.Sprintf("OpenClaw total: %d   Claude cost: $%.2f\n%s",
			last.OpenClawTotal, last.ClaudeCostUSD, spark,
		)
}

func sparkline(samples []openclaw.TokenSample) string {
	vals := make([]int64, 0, len(samples))
	for _, s := range samples {
		vals = append(vals, s.OpenClawTotal)
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	blocks := []rune("▁▂▃▄▅▆▇█")
	if max == min {
		return strings.Repeat(string(blocks[len(blocks)/2]), len(vals))
	}
	b := strings.Builder{}
	for _, v := range vals {
		r := float64(v-min) / float64(max-min)
		idx := int(r * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteRune(blocks[idx])
	}
	return b.String()
}

func renderSessions(sessions []openclaw.Session, subs []openclaw.SubagentRun, f sessionFilters) string {
	lines := []string{titleStyle.Render("Sessions / Subagents")}
	cut := time.Now().Add(-24 * time.Hour)
	for _, s := range sessions {
		if f.only24h && s.UpdatedAt.Before(cut) {
			continue
		}
		if f.hideRun && strings.Contains(s.Key, ":run:") {
			continue
		}
		if f.primaryModelOnly && f.primaryModel != "" && s.Model != f.primaryModel {
			continue
		}
		label := s.Label
		if label == "" {
			label = s.Key
		}
		lines = append(lines,
			fmt.Sprintf("%s  %s  %s",
				padRight(shortKey(s.Key), 28),
				padRight(firstN(label, 24), 24),
				padRight(modelShort(s.Model), 16),
			)+dimStyle.Render("  "+relTime(s.UpdatedAt)),
		)
	}
	if len(subs) > 0 {
		lines = append(lines, dimStyle.Render("subagents:"))
		for i, r := range subs {
			if i >= 6 {
				lines = append(lines, dimStyle.Render("…"))
				break
			}
			lines = append(lines, fmt.Sprintf("%s  %s", padRight(firstN(r.Label, 20), 20), dimStyle.Render(relTime(r.CreatedAt))))
		}
	}
	return strings.Join(lines, "\n")
}

func renderTasks(tasks []openclaw.Task, f taskFilters) string {
	lines := []string{titleStyle.Render("Latest Tasks")}
	flt := make([]openclaw.Task, 0, len(tasks))
	for _, t := range tasks {
		if f.levels != nil && !f.levels[t.Level] {
			continue
		}
		if f.sources != nil && !f.sources[t.Source] {
			continue
		}
		flt = append(flt, t)
	}
	sort.Slice(flt, func(i, j int) bool { return flt[i].At.After(flt[j].At) })
	if len(flt) == 0 {
		lines = append(lines, dimStyle.Render("(no tasks match filters)"))
		return strings.Join(lines, "\n")
	}
	for i, t := range flt {
		if i >= 20 {
			lines = append(lines, dimStyle.Render("…"))
			break
		}
		lvl := string(t.Level)
		src := string(t.Source)
		st := dimStyle
		switch t.Level {
		case openclaw.LevelError:
			st = badStyle
		case openclaw.LevelInfo:
			st = okStyle
		}
		lines = append(lines, fmt.Sprintf("%s %s %s %s", dimStyle.Render(timeFmt(t.At)), st.Render(padRight(lvl, 5)), dimStyle.Render(padRight(src, 8)), firstN(t.Title+": "+t.Detail, 80)))
	}
	return strings.Join(lines, "\n")
}

func renderCrons(crons []openclaw.CronJob) string {
	lines := []string{titleStyle.Render("Crons")}
	if len(crons) == 0 {
		lines = append(lines, dimStyle.Render("(no jobs.json)"))
		return strings.Join(lines, "\n")
	}
	for i, c := range crons {
		if i >= 12 {
			lines = append(lines, dimStyle.Render("…"))
			break
		}
		en := "on"
		if !c.Enabled {
			en = "off"
		}
		next := "-"
		if c.NextRun != nil {
			next = relTimeAbs(*c.NextRun)
		}
		last := "-"
		if c.LastRun != nil {
			last = relTime(*c.LastRun)
		}
		status := c.LastStatus
		st := dimStyle
		if status == "ok" {
			st = okStyle
		}
		if status == "error" {
			st = badStyle
		}
		err := firstN(c.LastError, 40)
		if err == "" {
			err = "-"
		}
		lines = append(lines, fmt.Sprintf("%s %s next:%s last:%s err:%s",
			padRight(firstN(c.Name, 20), 20),
			dimStyle.Render(en),
			next,
			last,
			st.Render(err),
		))
	}
	return strings.Join(lines, "\n")
}

func timeFmt(t time.Time) string { return t.Format("15:04:05") }

func relTimeAbs(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "due"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func shortKey(k string) string {
	if len(k) <= 28 {
		return k
	}
	return k[:27] + "…"
}

func modelShort(m string) string {
	if m == "" {
		return "-"
	}
	// show last path segment if any
	if idx := strings.LastIndex(m, "/"); idx >= 0 {
		m = m[idx+1:]
	}
	return m
}

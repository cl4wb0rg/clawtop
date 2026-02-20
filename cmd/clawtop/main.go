package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cl4wb0rg/clawtop/internal/openclaw"
	"github.com/cl4wb0rg/clawtop/internal/ui"
)

func main() {
	var (
		openclawRoot = flag.String("openclaw-root", "", "OpenClaw root dir (default: ~/.openclaw or $OPENCLAW_ROOT)")
		workspace    = flag.String("workspace", "", "Workspace dir (default: <openclaw-root>/workspace)")
		refresh      = flag.Duration("refresh", 2*time.Second, "refresh interval")
	)
	flag.Parse()

	paths, err := openclaw.DiscoverPaths(*openclawRoot, *workspace)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	m := ui.New(ui.Config{Paths: paths, Refresh: *refresh})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

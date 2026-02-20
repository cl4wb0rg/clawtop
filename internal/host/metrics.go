package host

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CPUStat struct {
	User, Nice, System, Idle, IOWait, IRQ, SoftIRQ, Steal uint64
}

type HostMetrics struct {
	At time.Time
	CPUPercent float64
	MemUsedBytes uint64
	MemTotalBytes uint64
	Load1 float64
	Load5 float64
	Load15 float64
}

func ReadCPUStat() (CPUStat, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return CPUStat{}, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return CPUStat{}, fmt.Errorf("/proc/stat empty")
	}
	fields := strings.Fields(s.Text())
	// cpu user nice system idle iowait irq softirq steal guest guest_nice
	if len(fields) < 9 {
		return CPUStat{}, fmt.Errorf("/proc/stat cpu line too short")
	}
	parse := func(i int) uint64 {
		x, _ := strconv.ParseUint(fields[i], 10, 64)
		return x
	}
	return CPUStat{User: parse(1), Nice: parse(2), System: parse(3), Idle: parse(4), IOWait: parse(5), IRQ: parse(6), SoftIRQ: parse(7), Steal: parse(8)}, nil
}

func CPUPercent(prev, cur CPUStat) float64 {
	prevIdle := prev.Idle + prev.IOWait
	curIdle := cur.Idle + cur.IOWait
	prevNon := prev.User + prev.Nice + prev.System + prev.IRQ + prev.SoftIRQ + prev.Steal
	curNon := cur.User + cur.Nice + cur.System + cur.IRQ + cur.SoftIRQ + cur.Steal
	prevTotal := prevIdle + prevNon
	curTotal := curIdle + curNon
	totald := float64(curTotal - prevTotal)
	idled := float64(curIdle - prevIdle)
	if totald <= 0 {
		return 0
	}
	return (totald - idled) / totald * 100.0
}

func ReadMemInfo() (total, available uint64, err error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	var memTotal, memAvail uint64
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "MemTotal:") {
			f := strings.Fields(ln)
			if len(f) >= 2 {
				memTotal, _ = strconv.ParseUint(f[1], 10, 64)
			}
		}
		if strings.HasPrefix(ln, "MemAvailable:") {
			f := strings.Fields(ln)
			if len(f) >= 2 {
				memAvail, _ = strconv.ParseUint(f[1], 10, 64)
			}
		}
	}
	// values are kB
	return memTotal * 1024, memAvail * 1024, nil
}

func ReadLoadAvg() (l1, l5, l15 float64, err error) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}
	f := strings.Fields(string(b))
	if len(f) < 3 {
		return 0, 0, 0, fmt.Errorf("/proc/loadavg too short")
	}
	l1, _ = strconv.ParseFloat(f[0], 64)
	l5, _ = strconv.ParseFloat(f[1], 64)
	l15, _ = strconv.ParseFloat(f[2], 64)
	return l1, l5, l15, nil
}

func HumanBytes(b uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1fG", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1fM", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1fK", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

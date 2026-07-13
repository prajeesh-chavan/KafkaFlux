package tui

import (
	"fmt"
	"strings"
	"time"
)

func dispatchCommand(cmd string, m *model) ([]string, bool) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil, false
	}

	switch parts[0] {
	case "/help", "/h":
		return []string{
			"Available commands:",
			"  /help, /h     Show this help",
			"  /profiles     List active profiles",
			"  /status       Show current metrics summary",
			"  /quit, /q     Shut down gracefully",
			"  Ctrl+C        Quit immediately",
		}, false

	case "/profiles", "/p":
		if len(m.profiles) == 0 {
			return []string{"No profiles loaded"}, false
		}
		lines := []string{fmt.Sprintf("Active profiles (%d):", len(m.profiles))}
		for _, name := range m.profiles {
			lines = append(lines, fmt.Sprintf("  - %s", name))
		}
		return lines, false

	case "/status", "/s":
		status := m.metrics.StatusJSON()
		lines := []string{
			fmt.Sprintf("Uptime: %s", formatDuration(timeSinceStart(m))),
			fmt.Sprintf("Buffer: %d/%d", safeInt64(status["buffer_used"]), safeInt64(status["buffer_capacity"])),
			fmt.Sprintf("Dropped: %d", safeInt64(status["events_dropped"])),
			fmt.Sprintf("Total Events: %d", safeInt64(status["total_events"])),
		}
		return lines, false

	case "/quit", "/q":
		return []string{"Initiating shutdown..."}, true

	default:
		return []string{fmt.Sprintf("Unknown command: %s (try /help)", parts[0])}, false
	}
}

func timeSinceStart(m *model) time.Duration {
	return time.Since(m.startTime)
}

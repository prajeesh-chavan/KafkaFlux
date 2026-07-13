package tui

import (
	"fmt"
	"strings"
	"time"

	"go-kafka-simulator/internal/telemetry"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg struct{}
type errMsg struct{ error }

type model struct {
	metrics    *telemetry.Metrics
	profiles   []string
	startTime  time.Time
	width      int
	height     int
	err        error
	quitting   bool
	cmdInput   textinput.Model
	messages   []string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#75B5AA"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	barStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87"))

	barEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C"))

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	cmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	msgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#87CEEB"))
)

func NewModel(m *telemetry.Metrics, profileNames []string) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "/help for commands"
	ti.Prompt = "> "
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 60

	return &model{
		metrics:   m,
		profiles:  profileNames,
		startTime: time.Now(),
		cmdInput:  ti,
		messages:  []string{"Type /help for available commands"},
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(tick(), textinput.Blink)
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			cmd := strings.TrimSpace(m.cmdInput.Value())
			m.cmdInput.SetValue("")
			if cmd != "" {
				messages, quit := dispatchCommand(cmd, m)
				m.messages = append(m.messages, messages...)
				if len(m.messages) > 100 {
					m.messages = m.messages[len(m.messages)-100:]
				}
				if quit {
					m.quitting = true
					return m, tea.Quit
				}
			}
			return m, nil
		case tea.KeyTab:
			m.cmdInput.Focus()
			return m, nil
		}

	case tickMsg:
		return m, tick()

	case errMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.cmdInput, cmd = m.cmdInput.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return "Shutting down...\n"
	}

	s := strings.Builder{}

	s.WriteString(titleStyle.Render(" KAFKAFLUX "))
	s.WriteString("\n\n")

	status := m.metrics.StatusJSON()
	uptime := int64(time.Since(m.startTime).Seconds())
	if u, ok := status["uptime_seconds"].(int64); ok {
		uptime = u
	}

	s.WriteString(labelStyle.Render("Uptime: "))
	s.WriteString(valueStyle.Render(formatDuration(time.Duration(uptime)*time.Second)))
	s.WriteString("   ")

	s.WriteString(labelStyle.Render("Buffer: "))
	bufUsed := safeInt64(status["buffer_used"])
	bufCap := safeInt64(status["buffer_capacity"])
	bar := renderBar(bufUsed, bufCap, 15)
	s.WriteString(bar)
	s.WriteString(fmt.Sprintf(" %d/%d", bufUsed, bufCap))
	s.WriteString("\n")

	s.WriteString(labelStyle.Render("Dropped: "))
	s.WriteString(valueStyle.Render(fmt.Sprintf("%d", safeInt64(status["events_dropped"]))))
	s.WriteString("   ")
	s.WriteString(labelStyle.Render("Failures: "))
	s.WriteString(valueStyle.Render(fmt.Sprintf("%d", safeInt64(status["delivery_failures"]))))
	s.WriteString("   ")
	s.WriteString(labelStyle.Render("Marshal Err: "))
	s.WriteString(valueStyle.Render(fmt.Sprintf("%d", safeInt64(status["marshal_errors"]))))
	s.WriteString("\n\n")

	s.WriteString(labelStyle.Render(fmt.Sprintf("%-20s %-12s %s", "Entity", "EPS", "Total Events")))
	s.WriteString("\n")
	s.WriteString(labelStyle.Render(strings.Repeat("─", 50)))
	s.WriteString("\n")

	if entities, ok := status["entities"].([]interface{}); ok {
		for _, e := range entities {
			if em, ok := e.(map[string]interface{}); ok {
				entity := safeString(em["entity"])
				eps := safeFloat64(em["eps"])
				events := safeInt64(em["events"])
				s.WriteString(fmt.Sprintf("%-20s %-12.1f %d\n", entity, eps, events))
			}
		}
	}
	s.WriteString("\n")

	for _, msg := range m.messages {
		s.WriteString(msgStyle.Render(msg))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(cmdStyle.Render(m.cmdInput.View()))

	return s.String()
}

func renderBar(current, capacity int64, width int) string {
	if capacity == 0 {
		return strings.Repeat(barEmptyStyle.Render("."), width)
	}
	filled := int(float64(current) / float64(capacity) * float64(width))
	if filled > width {
		filled = width
	}
	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteString(barStyle.Render("█"))
		} else {
			b.WriteString(barEmptyStyle.Render("."))
		}
	}
	return b.String()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func safeInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	}
	return 0
}

func safeFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	}
	return 0
}

func safeString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

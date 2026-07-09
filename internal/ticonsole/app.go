package ticonsole

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Tab is a Ti Console view.
type Tab int

const (
	TabDashboard Tab = iota
	TabServices
	TabHelp
)

// App is a read-only terminal control surface. Business logic remains behind
// Backend, which keeps the UI replaceable by upstream AI DevKit components.
type App struct {
	Backend         Backend
	In              io.Reader
	Out             io.Writer
	RefreshInterval time.Duration
	UseColor        bool
}

// Run starts the interactive dashboard. Commands are intentionally read-only:
// 1 dashboard, 2 services, 3 help, r refresh, q quit.
func (a *App) Run(ctx context.Context) error {
	if a.Backend == nil {
		return fmt.Errorf("ti console: backend is required")
	}
	if a.In == nil {
		a.In = os.Stdin
	}
	if a.Out == nil {
		a.Out = os.Stdout
	}
	if a.RefreshInterval <= 0 {
		a.RefreshInterval = 3 * time.Second
	}

	commands := make(chan rune, 1)
	errors := make(chan error, 1)
	go readCommands(a.In, commands, errors)

	ticker := time.NewTicker(a.RefreshInterval)
	defer ticker.Stop()

	currentTab := TabDashboard
	snapshot := a.Backend.Snapshot(ctx)
	a.render(currentTab, snapshot)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errors:
			if err == io.EOF {
				return nil
			}
			return err
		case command := <-commands:
			switch command {
			case 'q', 'Q':
				return nil
			case '1':
				currentTab = TabDashboard
			case '2':
				currentTab = TabServices
			case '3', '?':
				currentTab = TabHelp
			case 'r', 'R':
				snapshot = a.Backend.Snapshot(ctx)
			}
			a.render(currentTab, snapshot)
		case <-ticker.C:
			snapshot = a.Backend.Snapshot(ctx)
			a.render(currentTab, snapshot)
		}
	}
}

func readCommands(reader io.Reader, commands chan<- rune, errors chan<- error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		command, _ := utf8FirstRune(line)
		commands <- command
	}
	if err := scanner.Err(); err != nil {
		errors <- err
		return
	}
	errors <- io.EOF
}

func utf8FirstRune(value string) (rune, int) {
	for _, r := range value {
		return r, len(string(r))
	}
	return 0, 0
}

func (a *App) render(tab Tab, snapshot Snapshot) {
	fmt.Fprint(a.Out, "\x1b[H\x1b[2J")
	fmt.Fprint(a.Out, RenderSnapshot(tab, snapshot, a.UseColor))
}

// RenderSnapshot builds a deterministic terminal frame for tests and alternate
// renderers.
func RenderSnapshot(tab Tab, snapshot Snapshot, useColor bool) string {
	var output strings.Builder
	output.WriteString(style("Ti Console", "1;36", useColor))
	output.WriteString("  ")
	output.WriteString(style("AI DevKit-compatible control surface for Ti", "2", useColor))
	output.WriteString("\n")
	output.WriteString("Observed at: ")
	output.WriteString(snapshot.GeneratedAt.Format("2006-01-02 15:04:05"))
	output.WriteString("\n\n")

	output.WriteString(renderTabs(tab, useColor))
	output.WriteString("\n\n")

	switch tab {
	case TabServices:
		output.WriteString(renderServices(snapshot, true, useColor))
	case TabHelp:
		output.WriteString(renderHelp())
	default:
		output.WriteString(renderDashboard(snapshot, useColor))
	}

	output.WriteString("\n")
	output.WriteString(style("Commands", "1", useColor))
	output.WriteString(": 1 dashboard | 2 services | 3 help | r refresh | q quit")
	output.WriteString("\n")
	output.WriteString(style("Note", "1;33", useColor))
	output.WriteString(": enter a command then press Enter. Auto-refresh remains active.\n")
	return output.String()
}

func renderTabs(active Tab, useColor bool) string {
	tabs := []struct {
		label string
		tab   Tab
	}{
		{"[1] Dashboard", TabDashboard},
		{"[2] Services", TabServices},
		{"[3] Help", TabHelp},
	}
	parts := make([]string, 0, len(tabs))
	for _, item := range tabs {
		if item.tab == active {
			parts = append(parts, style(item.label, "1;7", useColor))
		} else {
			parts = append(parts, item.label)
		}
	}
	return strings.Join(parts, "  ")
}

func renderDashboard(snapshot Snapshot, useColor bool) string {
	var output strings.Builder
	output.WriteString(style("Runtime summary", "1", useColor))
	output.WriteString("\n")
	output.WriteString(fmt.Sprintf(
		"  healthy: %s   degraded: %s   down: %s   unknown: %s\n\n",
		style(fmt.Sprint(snapshot.Summary.Healthy), "1;32", useColor),
		style(fmt.Sprint(snapshot.Summary.Degraded), "1;33", useColor),
		style(fmt.Sprint(snapshot.Summary.Down), "1;31", useColor),
		style(fmt.Sprint(snapshot.Summary.Unknown), "1;37", useColor),
	))
	output.WriteString(renderServices(snapshot, false, useColor))
	if len(snapshot.Notices) > 0 {
		output.WriteString("\n")
		output.WriteString(style("Guardrails", "1", useColor))
		output.WriteString("\n")
		for _, notice := range snapshot.Notices {
			output.WriteString("  - ")
			output.WriteString(notice)
			output.WriteString("\n")
		}
	}
	return output.String()
}

func renderServices(snapshot Snapshot, detailed, useColor bool) string {
	var output strings.Builder
	if detailed {
		output.WriteString(style("Live service probes", "1", useColor))
		output.WriteString("\n")
	}
	output.WriteString(fmt.Sprintf("%-17s %-24s %-11s %-9s %s\n", "SERVICE", "ROLE", "STATE", "LATENCY", "ENDPOINT"))
	output.WriteString(strings.Repeat("-", 92))
	output.WriteString("\n")
	for _, service := range snapshot.Services {
		endpoint := service.Endpoint
		if endpoint == "" {
			endpoint = service.BaseURL
		}
		output.WriteString(fmt.Sprintf(
			"%-17s %-24s %-11s %-9s %s\n",
			truncate(service.Name, 17),
			truncate(service.Role, 24),
			stateLabel(service.State, useColor),
			durationMillis(service.Latency),
			endpoint,
		))
		if detailed && service.Detail != "" {
			output.WriteString("  detail: ")
			output.WriteString(service.Detail)
			if service.StatusCode != 0 {
				output.WriteString(fmt.Sprintf(" (HTTP %d)", service.StatusCode))
			}
			output.WriteString("\n")
		}
	}
	return output.String()
}

func renderHelp() string {
	return strings.Join([]string{
		"Ti Console MVP",
		"",
		"  - Read-only dashboard for TiBrain, Beads, 1MCP and Router candidates.",
		"  - Runtime status comes from live HTTP probes, never from static port labels.",
		"  - UI depends on a Backend interface so AI DevKit's renderer can replace it later.",
		"  - No start, stop, kill, delete or config mutation actions are enabled in this phase.",
		"",
		"Environment overrides:",
		"  TIBRAIN_URL      default http://127.0.0.1:1810",
		"  TI_BEADS_URL     default http://127.0.0.1:1811",
		"  TI_ROUTER_URLS   comma-separated candidate URLs",
		"",
	}, "\n")
}

func stateLabel(state ServiceState, useColor bool) string {
	label := strings.ToUpper(string(state))
	code := "1;37"
	switch state {
	case StateHealthy:
		code = "1;32"
	case StateDegraded:
		code = "1;33"
	case StateDown:
		code = "1;31"
	}
	return style(label, code, useColor)
}

func style(value, code string, enabled bool) string {
	if !enabled {
		return value
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func truncate(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

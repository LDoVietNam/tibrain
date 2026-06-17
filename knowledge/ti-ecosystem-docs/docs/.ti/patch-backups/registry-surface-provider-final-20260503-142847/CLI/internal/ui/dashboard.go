package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	// "github.com/ti/cli/internal/pluginhost/adapters"
	// "github.com/ti/pluginapi"
)

type DashboardItem struct {
	Title       string
	Command     string
	Description string
	Safety      string
	Icon        string
	Group       string
}

type dashboardModel struct {
	width    int
	height   int
	selected int
	message  string
	items    []DashboardItem
	filteredItems []DashboardItem
	filter   textinput.Model
	showFilter bool
	filtered bool
	wasFilterShown bool // Track filter state transition
}

// StartDashboardTUI launches the Ti mission-control terminal UI with plugin registry.
func StartDashboardTUI(registry any) error {
	// TODO: Integrate with new plugin registry
	// Get UI actions from plugin registry
	// uiActions := adapters.DashboardItems(registry)

	// Convert to DashboardItem format
	// items := make([]DashboardItem, 0, len(uiActions))
	// for _, action := range uiActions {
	// 	safety := "unknown"
	// 	// Determine safety from policy if available
	// 	items = append(items, DashboardItem{
	// 		Title:       action.Title,
	// 		Command:     action.Command,
	// 		Description: action.Description,
	// 		Safety:      safety,
	// 		Icon:        action.Icon,
	// 		Group:       action.Group,
	// 	})
	// }

	// If no plugin items, use fallback items
	// if len(items) == 0 {
	items := getFallbackItems()
	// }

	p := tea.NewProgram(dashboardModel{items: items}, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// StartDashboardTUILegacy launches the Ti mission-control terminal UI (legacy version without registry).
// This is kept for backward compatibility.
func StartDashboardTUILegacy() error {
	items := getFallbackItems()
	
	// Initialize filter input
	filter := textinput.New()
	filter.Placeholder = "Type / to filter commands..."
	filter.CharLimit = 50
	filter.Width = 30
	
	p := tea.NewProgram(dashboardModel{
		items:        items,
		filteredItems: items,
		filter:       filter,
		showFilter:   false,
		filtered:     false,
	}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// getFallbackItems returns the hardcoded fallback items
// This is now only used when the plugin registry is not available
// In production, all items should come from the plugin registry
func getFallbackItems() []DashboardItem {
	return []DashboardItem{
		// Core Workflows (these should be registered by the copilot plugin)
		{Title: "Review project", Command: "ti review --run", Description: "Auto-load skills, pack context, and review code/security.", Safety: "read-only", Icon: "🔍", Group: "Core"},
		{Title: "Fix build/test", Command: "ti fix --run", Description: "Run compressed diagnostics and produce a safe fix plan.", Safety: "asks before write", Icon: "🔧", Group: "Core"},
		{Title: "Plan work", Command: "ti plan \"describe your task\" --run", Description: "Create an implementation plan with context compression.", Safety: "read-only", Icon: "📋", Group: "Core"},
		{Title: "Optimize context", Command: "ti optimize context --run", Description: "Compact context and prepare useful handoff state.", Safety: "read-only", Icon: "⚡", Group: "Core"},

		// Engine Tools (new engines from Mega Engines Pack)
		{Title: "List runs", Command: "ti runs list", Description: "List all automation runs with status and metadata.", Safety: "read-only", Icon: "📊", Group: "Engines"},
		{Title: "Show memory", Command: "ti memory show", Description: "Show project memory entries with timestamps.", Safety: "read-only", Icon: "🧠", Group: "Engines"},
		{Title: "Scan secrets", Command: "ti secrets scan .", Description: "Scan codebase for potential secrets and sensitive data.", Safety: "read-only", Icon: "🔒", Group: "Engines"},
		{Title: "Task list", Command: "ti task list", Description: "List all tasks with status (todo/done).", Safety: "read-only", Icon: "✅", Group: "Engines"},
		{Title: "Next task", Command: "ti task next", Description: "Show the next pending task to work on.", Safety: "read-only", Icon: "➡️", Group: "Engines"},

		// Other Tools
		{Title: "Load skills folder", Command: "ti load-skill ./skills --run", Description: "Import your collected skill folder into .ti/skills.", Safety: "writes .ti only", Icon: "📚", Group: "Tools"},
		{Title: "Analyze repo", Command: "ti repo", Description: "Analyze GitHub repo with auto-clone, dependency graph, and pattern detection.", Safety: "read-only", Icon: "🔬", Group: "Tools"},
		{Title: "Use expert sub-agents", Command: "ti expert go \"review this repo\" --run", Description: "Route to planner/coder/reviewer/security/tester agents.", Safety: "read-only by default", Icon: "🤖", Group: "Tools"},
		{Title: "Open chat TUI", Command: "ti ask", Description: "Chat with the configured provider/router.", Safety: "provider call", Icon: "💬", Group: "Tools"},
		{Title: "Doctor", Command: "ti doctor", Description: "Check config, providers, skills, permissions, and repo health.", Safety: "read-only", Icon: "🏥", Group: "Tools"},
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseLeft:
			// Check if clicking on items list
			if msg.X < m.width-20 && msg.Y > 5 && msg.Y < 5+len(m.getCurrentItems()) {
				clickedIdx := msg.Y - 5
				items := m.getCurrentItems()
				if clickedIdx >= 0 && clickedIdx < len(items) {
					m.selected = clickedIdx
					item := items[m.selected]
					m.message = fmt.Sprintf("Run this command in your shell:\n\n  %s\n\nTi keeps this TUI safe: it shows the command instead of executing destructive actions from the menu.", item.Command)
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.showFilter {
				m.showFilter = false
				m.filtered = false
				m.filter.SetValue("")
				m.filter.Blur()
				m.selected = 0
				m.filteredItems = m.items
				return m, nil
			}
			return m, tea.Quit
		case "/":
			m.showFilter = true
			m.filter.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.showFilter {
				// Navigate filtered items when filter is shown
				if m.selected > 0 {
					m.selected--
				}
			} else {
				if m.selected > 0 {
					m.selected--
				}
			}
		case "down", "j":
			if m.showFilter {
				// Navigate filtered items when filter is shown
				items := m.getCurrentItems()
				if m.selected < len(items)-1 {
					m.selected++
				}
			} else {
				if m.selected < len(m.items)-1 {
					m.selected++
				}
			}
		case "enter":
			items := m.getCurrentItems()
			if m.selected < len(items) {
				item := items[m.selected]
				m.message = fmt.Sprintf("Run this command in your shell:\n\n  %s\n\nTi keeps this TUI safe: it shows the command instead of executing destructive actions from the menu.", item.Command)
			}
		case "?":
			m.message = "Keys: ↑/↓ or j/k to move, / to filter, Enter to show command, Esc to close filter/quit. Use `ti guide` for quick CLI help."
		}
	}
	
	// Update filter input only if shown and focused
	if m.showFilter {
		m.filter, cmd = m.filter.Update(msg)
		// Apply filter
		m.applyFilter()
		m.wasFilterShown = true
	} else if m.wasFilterShown {
		// Blur filter only when transitioning from shown to hidden
		m.filter.Blur()
		m.wasFilterShown = false
	}
	
	return m, tea.Batch(cmd)
}

func (m *dashboardModel) getCurrentItems() []DashboardItem {
	if m.filtered && len(m.filteredItems) > 0 {
		return m.filteredItems
	}
	return m.items
}

func (m *dashboardModel) applyFilter() {
	filterText := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	if filterText == "" {
		m.filtered = false
		m.filteredItems = m.items
		m.selected = 0
		return
	}
	
	var filtered []DashboardItem
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Title), filterText) ||
		   strings.Contains(strings.ToLower(item.Command), filterText) ||
		   strings.Contains(strings.ToLower(item.Description), filterText) ||
		   strings.Contains(strings.ToLower(item.Group), filterText) {
			filtered = append(filtered, item)
		}
	}
	
	m.filtered = true
	m.filteredItems = filtered
	m.selected = 0
}

func (m dashboardModel) View() string {
	if m.width == 0 {
		m.width = 100
	}
	w := maxInt(70, m.width-6)

	accent := lipgloss.Color("63")
	muted := lipgloss.Color("245")
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(accent).Padding(0, 2).Width(w).Render("✦ Ti Mission Control")
	
	// Update subtitle based on filter state
	subtitleText := "Không cần nhớ command: chọn việc cần làm, Ti sẽ gợi ý lệnh đúng. Enter = xem lệnh, q = thoát"
	if m.filtered {
		count := len(m.filteredItems)
		subtitleText = fmt.Sprintf("Filtered: %d/%d commands. / to change filter, Esc to clear", count, len(m.items))
	}
	subtitle := lipgloss.NewStyle().Foreground(muted).Render(subtitleText)

	// Add filter input if shown
	var filterSection string
	if m.showFilter {
		filterBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("99")).Padding(0, 1).Width(minInt(50, w)).Render(m.filter.View())
		filterSection = "\n" + filterBox
	}

	var rows []string
	items := m.getCurrentItems()
	for i, item := range items {
		cursor := "  "
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == m.selected {
			cursor = "› "
			style = style.Foreground(lipgloss.Color("230")).Background(accent).Bold(true)
		}
		left := fmt.Sprintf("%s%-24s", cursor, item.Title)
		cmd := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(item.Command)
		safe := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(item.Safety)
		line := style.Render(left) + "  " + cmd + "  " + safe
		rows = append(rows, line)
	}

	detail := items[m.selected]
	detailBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(1, 2).Width(w).Render(
		lipgloss.NewStyle().Bold(true).Render(detail.Title) + "\n" +
			detail.Description + "\n\nCommand:\n  " + detail.Command,
	)

	msg := m.message
	if strings.TrimSpace(msg) == "" {
		msg = "Tip: nếu bạn không biết dùng gì, chạy `ti guide` hoặc gõ tự nhiên: `ti \"sua loi build\"`. Type / to filter commands."
	}
	msgBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).Padding(1, 2).Width(w).Render(msg)

	return lipgloss.NewStyle().Margin(1, 2).Render(title + "\n" + subtitle + filterSection + "\n\n" + strings.Join(rows, "\n") + "\n\n" + detailBox + "\n" + msgBox)
}

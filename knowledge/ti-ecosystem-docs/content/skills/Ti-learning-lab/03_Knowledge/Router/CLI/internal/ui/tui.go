package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ti/cli/internal/providers"
)

// ChatMessage represents a single message in the chat.
type ChatMessage struct {
	Role    string
	Content string
	Time    time.Time
}

// ChatModel is the Bubble Tea model for the chat TUI.
type ChatModel struct {
	width       int
	height      int
	messages    []ChatMessage
	input       textinput.Model
	provider    providers.Provider
	registry    *providers.Registry
	selectedIdx int
	providers   []string
	err         error
	loading     bool
}

// StartChatTUI launches the interactive chat TUI.
func StartChatTUI(reg *providers.Registry) error {
	m := initialChatModel(reg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialChatModel(reg *providers.Registry) ChatModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 4000
	ti.Width = 80

	plist := reg.List()
	selected := ""
	if len(plist) > 0 {
		selected = plist[0]
	}

	var p providers.Provider
	if selected != "" {
		p, _ = reg.Get(selected)
	}

	return ChatModel{
		input:       ti,
		provider:    p,
		registry:    reg,
		providers:   plist,
		selectedIdx: 0,
		messages:    []ChatMessage{},
	}
}

// Init implements tea.Model.
func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			if m.selectedIdx > 0 && len(m.providers) > 0 {
				m.selectedIdx--
				m.provider, _ = m.registry.Get(m.providers[m.selectedIdx])
			}
		case tea.KeyDown:
			if m.selectedIdx < len(m.providers)-1 {
				m.selectedIdx++
				m.provider, _ = m.registry.Get(m.providers[m.selectedIdx])
			}
		case tea.KeyEnter:
			if m.input.Value() == "" || m.provider == nil || m.loading {
				return m, nil
			}
			prompt := m.input.Value()
			m.messages = append(m.messages, ChatMessage{
				Role:    "user",
				Content: prompt,
				Time:    time.Now(),
			})
			m.input.SetValue("")
			m.loading = true
			m.err = nil
			return m, m.sendMessage(prompt)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 4

	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.messages = append(m.messages, ChatMessage{
				Role:    "error",
				Content: msg.err.Error(),
				Time:    time.Now(),
			})
		} else {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: msg.content,
				Time:    time.Now(),
			})
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m ChatModel) View() string {
	if m.width == 0 {
		m.width = 80
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render("Ti Chat"))
	b.WriteString("\n")

	// Provider selector
	providerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if len(m.providers) > 0 {
		b.WriteString(providerStyle.Render(fmt.Sprintf("Provider: %s | ↑↓ to switch | ESC to quit", m.providerName())))
	} else {
		b.WriteString(providerStyle.Render("No providers available. Run 'ti provider list' to check."))
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width-2))
	b.WriteString("\n")

	// Messages
	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			userStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
			b.WriteString(userStyle.Render("You: "))
			b.WriteString(msg.Content)
		case "assistant":
			aiStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
			b.WriteString(aiStyle.Render("AI: "))
			b.WriteString(msg.Content)
		case "error":
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			b.WriteString(errStyle.Render("Error: " + msg.Content))
		}
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("Thinking..."))
		b.WriteString("\n")
	}

	// Input area
	b.WriteString(strings.Repeat("─", m.width-2))
	b.WriteString("\n")
	b.WriteString(m.input.View())

	return b.String()
}

func (m ChatModel) providerName() string {
	if m.provider == nil {
		return "none"
	}
	return m.provider.Name()
}

type responseMsg struct {
	content string
	err     error
}

func (m ChatModel) sendMessage(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		req := providers.ChatRequest{
			Model:    "",
			Messages: buildMessages(m.messages),
		}

		resp, err := m.provider.Chat(ctx, req)
		if err != nil {
			return responseMsg{err: err}
		}
		return responseMsg{content: resp.Content}
	}
}

func buildMessages(msgs []ChatMessage) []providers.Message {
	out := make([]providers.Message, len(msgs))
	for i, m := range msgs {
		out[i] = providers.Message{Role: m.Role, Content: m.Content}
	}
	return out
}

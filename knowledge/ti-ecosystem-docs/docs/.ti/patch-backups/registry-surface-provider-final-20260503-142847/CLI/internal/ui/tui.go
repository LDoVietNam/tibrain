package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	// "github.com/ti/cli/internal/pluginhost/adapters"
	"github.com/ti/cli/internal/providers"
	// "github.com/ti/pluginapi"
)

// ChatMessage represents a single message in the chat.
type ChatMessage struct {
	Role    string
	Content string
	Time    time.Time
}

// ChatModel is the Bubble Tea model for the chat TUI.
type ChatModel struct {
	width         int
	height        int
	messages      []ChatMessage
	input         textinput.Model
	provider      providers.Provider
	registry      *providers.Registry
	pluginRegistry any
	selectedIdx   int
	providers     []string
	err           error
	loading       bool
	scroll        int
	showHelp      bool
	status        string
	showSuggestions bool
	suggestions   []string
	selectedSuggestion int
	selectedCommandIdx int
	commandList    []string
	focusMode     string // "input", "sidebar", or "form"
	showProviderForm bool
	providerFormName textinput.Model
	providerFormBaseURL textinput.Model
	providerFormAPIKey textinput.Model
	providerFormFocus int // 0=name, 1=baseurl, 2=apikey
}

// StartChatTUI launches the interactive chat TUI.
func StartChatTUI(reg *providers.Registry, pluginReg any) error {
	m := initialChatModel(reg, pluginReg)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// StartChatTUILegacy launches the interactive chat TUI (legacy version without plugin registry).
func StartChatTUILegacy(reg *providers.Registry) error {
	return StartChatTUI(reg, nil)
}

func initialChatModel(reg *providers.Registry, pluginReg any) ChatModel {
	ti := textinput.New()
	ti.Placeholder = "Ask Ti...  Type / for commands"
	ti.Focus()
	ti.CharLimit = 12000
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

	// Get slash suggestions from plugin registry
	suggestions := []string{"/help", "/clear", "/providers", "/provider", "/model", "/add-provider", "/exit"}
	commandList := []string{"/help", "/clear", "/providers", "/provider", "/model", "/add-provider", "/exit"}
	if pluginReg != nil {
		// TODO: Integrate with new plugin registry
		// suggestions = adapters.SlashSuggestions(pluginReg, "/")
		// commandList = adapters.CommandList(pluginReg)
	}

	// Initialize provider form inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "Provider name"
	nameInput.CharLimit = 50
	nameInput.Width = 30

	baseURLInput := textinput.New()
	baseURLInput.Placeholder = "https://api.example.com/v1"
	baseURLInput.CharLimit = 200
	baseURLInput.Width = 40

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "API key (or leave empty for no-key)"
	apiKeyInput.CharLimit = 200
	apiKeyInput.Width = 40
	apiKeyInput.EchoMode = textinput.EchoPassword

	return ChatModel{
		input:              ti,
		provider:           p,
		registry:           reg,
		pluginRegistry:     pluginReg,
		providers:          plist,
		selectedIdx:        0,
		showSuggestions:   false,
		suggestions:       suggestions,
		selectedSuggestion: 0,
		selectedCommandIdx: 0,
		commandList:        commandList,
		focusMode:         "input",
		showProviderForm: false,
		providerFormName: nameInput,
		providerFormBaseURL: baseURLInput,
		providerFormAPIKey: apiKeyInput,
		providerFormFocus: 0,
		messages: []ChatMessage{{
			Role:    "system",
			Content: "Ti Chat ready. Type / for commands, Tab to switch provider, Ctrl+L to clear, PgUp/PgDn to scroll. Use Tab to switch focus between input and sidebar. Click commands or suggestions to select them.",
			Time:    time.Now(),
		}},
		status: "ready",
	}
}

// Init implements tea.Model.
func (m ChatModel) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		// Handle mouse clicks
		switch msg.Type {
		case tea.MouseLeft:
			// Check if clicking on suggestions
			if m.showSuggestions && len(m.suggestions) > 0 {
				// Simple heuristic: if click is in suggestion area (bottom of screen)
				if msg.Y > m.height-10-len(m.suggestions) {
					clickedIdx := msg.Y - (m.height - 10 - len(m.suggestions))
					if clickedIdx >= 0 && clickedIdx < len(m.suggestions) {
						m.selectedSuggestion = clickedIdx
						m.input.SetValue(m.suggestions[m.selectedSuggestion] + " ")
						m.showSuggestions = false
						m.selectedSuggestion = 0
						m.focusMode = "input"
						m.input.Focus()
						var cmd tea.Cmd
						m.input, cmd = m.input.Update(tea.KeyMsg{Type: tea.KeyEnter})
						return m, cmd
					}
				}
			}
			// Check if clicking on sidebar commands
			if msg.X < 35 && msg.Y > 10 && msg.Y < 10+len(m.commandList) {
				clickedIdx := msg.Y - 10
				if clickedIdx >= 0 && clickedIdx < len(m.commandList) {
					m.selectedCommandIdx = clickedIdx
					cmd := m.commandList[m.selectedCommandIdx]
					// If it's /add-provider, open form
					if cmd == "/add-provider" {
						m.showProviderForm = true
						m.focusMode = "form"
						m.providerFormFocus = 0
						m.providerFormName.Focus()
						m.input.Blur()
						return m, nil
					}
					// Otherwise, fill input
					m.input.SetValue(cmd + " ")
					m.focusMode = "input"
					m.input.Focus()
					var teaCmd tea.Cmd
					m.input, teaCmd = m.input.Update(tea.KeyMsg{Type: tea.KeyEnter})
					return m, teaCmd
				}
			}
		}
	case tea.KeyMsg:
		// Handle suggestions
		if m.showSuggestions {
			switch msg.String() {
			case "esc":
				m.showSuggestions = false
				m.selectedSuggestion = 0
				return m, nil
			case "up":
				if m.selectedSuggestion > 0 {
					m.selectedSuggestion--
				}
				return m, nil
			case "down":
				if m.selectedSuggestion < len(m.suggestions)-1 {
					m.selectedSuggestion++
				}
				return m, nil
			case "enter":
				if len(m.suggestions) > 0 && m.selectedSuggestion < len(m.suggestions) {
					m.input.SetValue(m.suggestions[m.selectedSuggestion] + " ")
					m.showSuggestions = false
					m.selectedSuggestion = 0
					var cmd tea.Cmd
					m.input, cmd = m.input.Update(msg)
					return m, cmd
				}
			case "tab":
				m.showSuggestions = false
				m.selectedSuggestion = 0
				m.nextProvider()
				return m, nil
			}
		}

		// Handle normal input
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.showProviderForm {
				m.showProviderForm = false
				m.focusMode = "input"
				m.input.Focus()
				return m, nil
			}
			if m.focusMode == "sidebar" {
				m.focusMode = "input"
				m.input.Focus()
				return m, nil
			}
			return m, tea.Quit
		case "tab":
			if m.showProviderForm {
				// Switch between form fields
				m.providerFormFocus = (m.providerFormFocus + 1) % 3
				switch m.providerFormFocus {
				case 0:
					m.providerFormName.Focus()
					m.providerFormBaseURL.Blur()
					m.providerFormAPIKey.Blur()
				case 1:
					m.providerFormName.Blur()
					m.providerFormBaseURL.Focus()
					m.providerFormAPIKey.Blur()
				case 2:
					m.providerFormName.Blur()
					m.providerFormBaseURL.Blur()
					m.providerFormAPIKey.Focus()
				}
				return m, nil
			}
			if m.showSuggestions {
				// Navigate suggestions
				if m.selectedSuggestion < len(m.suggestions)-1 {
					m.selectedSuggestion++
				} else {
					m.selectedSuggestion = 0
				}
				return m, nil
			}
			// Switch focus between input and sidebar
			if m.focusMode == "input" {
				if strings.HasPrefix(m.input.Value(), "/") {
					m.updateSuggestions()
					m.showSuggestions = true
					return m, nil
				}
				m.focusMode = "sidebar"
				m.input.Blur()
			} else {
				m.focusMode = "input"
				m.input.Focus()
			}
			return m, nil
		case "shift+tab":
			if m.focusMode == "sidebar" {
				m.focusMode = "input"
				m.input.Focus()
			} else {
				m.nextProvider()
			}
			return m, nil
		case "up":
			if m.focusMode == "sidebar" {
				if m.selectedCommandIdx > 0 {
					m.selectedCommandIdx--
				}
				return m, nil
			}
		case "down":
			if m.focusMode == "sidebar" {
				if m.selectedCommandIdx < len(m.commandList)-1 {
					m.selectedCommandIdx++
				}
				return m, nil
			}
		case "ctrl+l":
			m.messages = nil
			m.scroll = 0
			m.status = "cleared"
		case "pgup":
			m.scroll = minInt(len(m.messages), m.scroll+1)
		case "pgdown":
			m.scroll = maxInt(0, m.scroll-1)
		case "?":
			m.showHelp = !m.showHelp
		case "enter":
			if m.showProviderForm {
				// Submit provider form
				name := strings.TrimSpace(m.providerFormName.Value())
				baseURL := strings.TrimSpace(m.providerFormBaseURL.Value())
				apiKey := strings.TrimSpace(m.providerFormAPIKey.Value())
				
				if name == "" || baseURL == "" {
					m.messages = append(m.messages, ChatMessage{Role: "error", Content: "Provider name and base URL are required.", Time: time.Now()})
					return m, nil
				}
				
				// Save provider
				noKey := apiKey == ""
				saved := providers.SavedCompatibleProvider{
					Name:      name,
					BaseURL:   baseURL,
					Model:     "auto",
					Models:    []string{"auto"},
					APIKey:    apiKey,
					NoKey:     noKey,
				}
				
				path, err := providers.SaveCustomCompatibleProvider(saved)
				if err != nil {
					m.messages = append(m.messages, ChatMessage{Role: "error", Content: fmt.Sprintf("Failed to save provider: %v", err), Time: time.Now()})
					return m, nil
				}
				
				// Try to load and register
				p, err := providers.NewCompatibleProvider(providers.CompatibleProviderConfig{
					Name:         name,
					BaseURL:      baseURL,
					APIKey:       apiKey,
					DefaultModel: "auto",
					Models:       []string{"auto"},
					HealthyNoKey: noKey,
				})
				
				if err == nil {
					_ = m.registry.Register(p)
					m.providers = m.registry.List()
				}
				
				// Clear form and close
				m.providerFormName.SetValue("")
				m.providerFormBaseURL.SetValue("")
				m.providerFormAPIKey.SetValue("")
				m.showProviderForm = false
				m.focusMode = "input"
				m.input.Focus()
				
				msg := fmt.Sprintf("✅ Provider saved: %s\n", name)
				msg += fmt.Sprintf("Store: %s\n", path)
				msg += fmt.Sprintf("Base URL: %s\n", baseURL)
				if noKey {
					msg += "Auth: no-key/local\n"
				} else if apiKey != "" {
					msg += "Auth: saved-key\n"
				}
				msg += "\nProvider is now available. Use /providers to list all providers."
				
				m.messages = append(m.messages, ChatMessage{Role: "system", Content: msg, Time: time.Now()})
				return m, nil
			}
			if m.showSuggestions {
				if len(m.suggestions) > 0 && m.selectedSuggestion < len(m.suggestions) {
					m.input.SetValue(m.suggestions[m.selectedSuggestion] + " ")
					m.showSuggestions = false
					m.selectedSuggestion = 0
					m.focusMode = "input"
					m.input.Focus()
					var cmd tea.Cmd
					m.input, cmd = m.input.Update(msg)
					return m, cmd
				}
			}
			if m.focusMode == "sidebar" {
				// Execute selected command from sidebar
				if m.selectedCommandIdx < len(m.commandList) {
					cmd := m.commandList[m.selectedCommandIdx]
					// If it's /add-provider, open form
					if cmd == "/add-provider" {
						m.showProviderForm = true
						m.focusMode = "form"
						m.providerFormFocus = 0
						m.providerFormName.Focus()
						m.input.Blur()
						return m, nil
					}
					m.input.SetValue(cmd + " ")
					m.focusMode = "input"
					m.input.Focus()
					m.showSuggestions = false
					var teaCmd tea.Cmd
					m.input, teaCmd = m.input.Update(msg)
					return m, teaCmd
				}
				return m, nil
			}
			if m.loading {
				return m, nil
			}
			prompt := strings.TrimSpace(m.input.Value())
			if prompt == "" {
				return m, nil
			}
			m.input.SetValue("")
			m.showSuggestions = false
			if handled, cmd := m.handleSlashCommand(prompt); handled {
				return m, cmd
			}
			if m.provider == nil {
				m.messages = append(m.messages, ChatMessage{Role: "error", Content: "No provider is available. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or start Ti Router.", Time: time.Now()})
				return m, nil
			}
			m.messages = append(m.messages, ChatMessage{Role: "user", Content: prompt, Time: time.Now()})
			m.loading = true
			m.err = nil
			m.status = "sending to " + m.providerName()
			return m, m.sendMessage(prompt)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = maxInt(20, msg.Width-10)

	case responseMsg:
		m.loading = false
		m.status = "ready"
		if msg.err != nil {
			m.err = msg.err
			m.messages = append(m.messages, ChatMessage{Role: "error", Content: msg.err.Error(), Time: time.Now()})
		} else {
			m.messages = append(m.messages, ChatMessage{Role: "assistant", Content: msg.content, Time: time.Now()})
		}
	}

	var cmd tea.Cmd
	
	// Update form inputs if in form mode
	if m.showProviderForm {
		switch m.providerFormFocus {
		case 0:
			m.providerFormName, cmd = m.providerFormName.Update(msg)
		case 1:
			m.providerFormBaseURL, cmd = m.providerFormBaseURL.Update(msg)
		case 2:
			m.providerFormAPIKey, cmd = m.providerFormAPIKey.Update(msg)
		}
		cmds = append(cmds, cmd)
	} else {
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		
		// Auto-show suggestions when typing /
		currentInput := m.input.Value()
		if strings.HasPrefix(currentInput, "/") {
			m.updateSuggestions()
			// Always show suggestions when typing / (unless space after)
			if !strings.HasSuffix(currentInput, " ") {
				m.showSuggestions = true
			} else {
				m.showSuggestions = false
			}
		} else {
			m.showSuggestions = false
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) updateSuggestions() {
	input := strings.TrimSpace(m.input.Value())

	// Use plugin registry if available
	if m.pluginRegistry != nil {
		// TODO: Integrate with new plugin registry
		// m.suggestions = adapters.SlashSuggestions(m.pluginRegistry, input)
		// m.selectedSuggestion = 0
		// return
	}

	// Fallback to hardcoded suggestions (built-in commands only)
	inputLower := strings.ToLower(input)

	// Show all commands when just / or empty
	if input == "/" || input == "" {
		m.suggestions = []string{"/help", "/clear", "/providers", "/provider", "/model", "/add-provider", "/exit"}
		m.selectedSuggestion = 0
		return
	}

	// Filter commands that start with input
	var filtered []string
	for _, cmd := range []string{"/help", "/clear", "/providers", "/provider", "/model", "/add-provider", "/exit"} {
		if strings.HasPrefix(cmd, inputLower) {
			filtered = append(filtered, cmd)
		}
	}

	if len(filtered) > 0 {
		m.suggestions = filtered
		m.selectedSuggestion = 0
	} else {
		m.suggestions = []string{}
	}
}

func (m *ChatModel) nextProvider() {
	if len(m.providers) == 0 {
		return
	}
	m.selectedIdx = (m.selectedIdx + 1) % len(m.providers)
	m.provider, _ = m.registry.Get(m.providers[m.selectedIdx])
	m.status = "provider: " + m.providerName()
}

func (m *ChatModel) prevProvider() {
	if len(m.providers) == 0 {
		return
	}
	m.selectedIdx--
	if m.selectedIdx < 0 {
		m.selectedIdx = len(m.providers) - 1
	}
	m.provider, _ = m.registry.Get(m.providers[m.selectedIdx])
	m.status = "provider: " + m.providerName()
}

func (m *ChatModel) handleSlashCommand(prompt string) (bool, tea.Cmd) {
	if !strings.HasPrefix(prompt, "/") {
		return false, nil
	}
	fields := strings.Fields(prompt)
	cmdName := fields[0]

	// Handle built-in commands first
	cmd := strings.ToLower(strings.TrimPrefix(cmdName, "/"))
	switch cmd {
	case "q", "quit", "exit":
		return true, tea.Quit
	case "clear", "cls":
		m.messages = nil
		m.status = "cleared"
		return true, nil
	case "help", "?":
		m.showHelp = !m.showHelp
		return true, nil
	case "providers":
		m.messages = append(m.messages, ChatMessage{Role: "system", Content: m.providerListText(), Time: time.Now()})
		return true, nil
	case "provider":
		if len(fields) < 2 {
			m.messages = append(m.messages, ChatMessage{Role: "system", Content: "Usage: /provider <name-or-number>", Time: time.Now()})
			return true, nil
		}
		m.selectProvider(fields[1])
		return true, nil
	case "model":
		m.messages = append(m.messages, ChatMessage{Role: "system", Content: m.modelInfo(), Time: time.Now()})
		return true, nil
	case "add-provider":
		// Auto-open form for better UX
		m.showProviderForm = true
		m.focusMode = "form"
		m.providerFormFocus = 0
		m.providerFormName.Focus()
		m.input.Blur()
		m.input.SetValue("")
		return true, nil
	}

	// Try plugin registry for other commands
	if m.pluginRegistry != nil {
		// TODO: Integrate with new plugin registry
		// slashHandler := adapters.NewSlashCommandHandler(m.pluginRegistry)
		// ctx := map[string]any{}
		// args := []string{}
		// if len(fields) > 1 {
		// 	args = fields[1:]
		// }
		// result, err := slashHandler.HandleSlashCommand(cmdName, args, ctx)
		// if err != nil {
		// 	m.messages = append(m.messages, ChatMessage{Role: "error", Content: err.Error(), Time: time.Now()})
		// } else if result.Text != "" {
		// 	m.messages = append(m.messages, ChatMessage{Role: "system", Content: result.Text, Time: time.Now()})
		// }
		// return true, nil
	}

	// Unknown command
	m.messages = append(m.messages, ChatMessage{Role: "system", Content: "Unknown command. Try /help, /providers, /provider <name>, /model, /add-provider, /clear, /exit.", Time: time.Now()})
	return true, nil
}

func (m *ChatModel) modelInfo() string {
	if m.provider == nil {
		return "No provider selected. Use /provider <name> to select one."
	}
	
	info := fmt.Sprintf("Provider: %s\n", m.provider.Name())
	info += fmt.Sprintf("Default Model: %s\n", m.provider.DefaultModel())
	
	models := m.provider.Models()
	if len(models) > 0 {
		info += fmt.Sprintf("Available Models (%d):\n", len(models))
		for i, model := range models {
			info += fmt.Sprintf("  %d. %s\n", i+1, model)
		}
	} else {
		info += "No specific model list available (provider may support dynamic models)"
	}
	
	return info
}

func (m *ChatModel) addProvider(name, baseURL, model string, extraArgs []string) {
	name = strings.TrimSpace(name)
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	model = strings.TrimSpace(model)
	
	if name == "" || baseURL == "" {
		m.messages = append(m.messages, ChatMessage{Role: "error", Content: "Provider name and base URL are required.", Time: time.Now()})
		return
	}
	
	if model == "" {
		model = "auto"
	}
	
	// Handle API key
	apiKey := ""
	noKey := false
	if len(extraArgs) > 0 {
		arg := strings.TrimSpace(extraArgs[0])
		if strings.EqualFold(arg, "nokey") || strings.EqualFold(arg, "no-key") {
			noKey = true
		} else if strings.HasPrefix(strings.ToLower(arg), "env:") {
			envVar := strings.TrimSpace(arg[4:])
			apiKey = os.Getenv(envVar)
			if apiKey == "" {
				m.messages = append(m.messages, ChatMessage{Role: "error", Content: fmt.Sprintf("Environment variable %s is not set or empty.", envVar), Time: time.Now()})
				return
			}
		} else {
			apiKey = arg
		}
	}
	
	// Save provider
	saved := providers.SavedCompatibleProvider{
		Name:      name,
		BaseURL:   baseURL,
		Model:     model,
		Models:    []string{model},
		APIKey:    apiKey,
		NoKey:     noKey,
	}
	
	path, err := providers.SaveCustomCompatibleProvider(saved)
	if err != nil {
		m.messages = append(m.messages, ChatMessage{Role: "error", Content: fmt.Sprintf("Failed to save provider: %v", err), Time: time.Now()})
		return
	}
	
	// Try to load and register the provider
	var providerAPIKey string
	if apiKey != "" {
		providerAPIKey = apiKey
	} else if !noKey && strings.HasPrefix(strings.TrimSpace(extraArgs[0]), "env:") {
		envVar := strings.TrimSpace(extraArgs[0][4:])
		providerAPIKey = os.Getenv(envVar)
	}
	
	p, err := providers.NewCompatibleProvider(providers.CompatibleProviderConfig{
		Name:         name,
		BaseURL:      baseURL,
		APIKey:       providerAPIKey,
		DefaultModel: model,
		Models:       []string{model},
		HealthyNoKey: noKey,
	})
	
	if err == nil {
		_ = m.registry.Register(p)
		m.providers = m.registry.List()
		m.status = "provider added: " + name
	} else {
		m.status = "provider saved but may not be healthy"
	}
	
	msg := fmt.Sprintf("✅ Provider saved: %s\n", name)
	msg += fmt.Sprintf("Store: %s\n", path)
	msg += fmt.Sprintf("Base URL: %s\n", baseURL)
	msg += fmt.Sprintf("Model: %s\n", model)
	if noKey {
		msg += "Auth: no-key/local\n"
	} else if strings.HasPrefix(strings.TrimSpace(extraArgs[0]), "env:") {
		msg += fmt.Sprintf("Auth: env:%s\n", strings.TrimSpace(extraArgs[0][4:]))
	} else if apiKey != "" {
		msg += "Auth: saved-key\n"
	}
	msg += "\nProvider is now available. Use /providers to list all providers."
	
	m.messages = append(m.messages, ChatMessage{Role: "system", Content: msg, Time: time.Now()})
}

func (m *ChatModel) selectProvider(token string) {
	if len(m.providers) == 0 {
		m.status = "no providers"
		return
	}
	for i, name := range m.providers {
		if fmt.Sprint(i+1) == token || strings.EqualFold(name, token) {
			m.selectedIdx = i
			m.provider, _ = m.registry.Get(name)
			m.status = "provider: " + name
			m.messages = append(m.messages, ChatMessage{Role: "system", Content: "Provider switched to " + name, Time: time.Now()})
			return
		}
	}
	m.messages = append(m.messages, ChatMessage{Role: "error", Content: "Provider not found: " + token, Time: time.Now()})
}

// View implements tea.Model.
func (m ChatModel) View() string {
	if m.width == 0 {
		m.width = 110
	}
	if m.height == 0 {
		m.height = 34
	}
	outerW := maxInt(76, m.width-4)
	sidebarW := minInt(34, maxInt(26, outerW/4))
	chatW := maxInt(44, outerW-sidebarW-3)

	header := RenderLogo(outerW)
	sidebar := m.renderChatSidebar(sidebarW)
	chat := m.renderChatPanel(chatW)
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, "  ", chat)
	footer := m.renderChatFooter(outerW)
	return lipgloss.NewStyle().Margin(1, 2).Render(header + "\n" + body + "\n" + footer)
}

func (m ChatModel) renderChatSidebar(width int) string {
	var rows []string
	rows = append(rows, badge("CHAT", colorInk, colorViolet))
	rows = append(rows, "")
	if len(m.providers) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(colorRed).Render("No providers"))
		rows = append(rows, lipgloss.NewStyle().Foreground(colorMuted).Render("Set OPENAI_API_KEY or OPENROUTER_API_KEY."))
	} else {
		rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(colorInk).Render("Providers"))
		for i, name := range m.providers {
			p, _ := m.registry.Get(name)
			state := "●"
			stateColor := colorGreen
			if p == nil || !p.IsHealthy() {
				stateColor = colorAmber
			}
			line := fmt.Sprintf("%s %d. %s", lipgloss.NewStyle().Foreground(stateColor).Render(state), i+1, name)
			style := lipgloss.NewStyle().Width(width-4).Padding(0, 1)
			if i == m.selectedIdx {
				style = style.Bold(true).Foreground(colorInk).Background(colorBlue)
			}
			rows = append(rows, style.Render(truncateText(line, width-5)))
		}
	}
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(colorInk).Render("Commands"))
	for i, line := range m.commandList {
		style := lipgloss.NewStyle().Foreground(colorMuted)
		if i == m.selectedCommandIdx && m.focusMode == "sidebar" {
			style = style.Bold(true).Foreground(colorInk).Background(colorBlue)
		}
		rows = append(rows, style.Render(line))
	}
	rows = append(rows, "")
	if m.focusMode == "sidebar" {
		rows = append(rows, lipgloss.NewStyle().Foreground(colorCyan).Render("↑↓ select  Enter execute  Tab back"))
	} else {
		rows = append(rows, lipgloss.NewStyle().Foreground(colorSubtle).Render("Tab to select commands"))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorPanelHi).Padding(1, 2).Width(width).Render(strings.Join(rows, "\n"))
}

func (m ChatModel) renderChatPanel(width int) string {
	availableH := maxInt(8, m.height-18)
	visible := m.visibleMessages(availableH)
	var parts []string
	if len(visible) == 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(colorMuted).Italic(true).Render("Start chatting below. Type / for commands."))
	} else {
		for _, msg := range visible {
			parts = append(parts, renderMessage(msg, width-6))
		}
	}
	if m.loading {
		parts = append(parts, lipgloss.NewStyle().Foreground(colorAmber).Render("⏳ Ti is thinking..."))
	}
	if m.showHelp {
		parts = append(parts, m.helpText(width-6))
	}
	messages := strings.TrimRight(strings.Join(parts, "\n"), "\n")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorViolet).Padding(1, 2).Width(width).Height(availableH).Render(messages)
	inputBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorBlue).Padding(0, 1).Width(width).Render(m.input.View())
	
	// Add provider form modal if showing
	if m.showProviderForm {
		formBox := m.renderProviderForm(width)
		return box + "\n" + inputBox + "\n" + formBox
	}
	
	// Add suggestions popup if showing
	if m.showSuggestions && len(m.suggestions) > 0 {
		suggestionBox := m.renderSuggestions(width)
		return box + "\n" + inputBox + "\n" + suggestionBox
	}
	
	return box + "\n" + inputBox
}

func (m ChatModel) renderSuggestions(width int) string {
	var rows []string
	for i, suggestion := range m.suggestions {
		prefix := " "
		if i == m.selectedSuggestion {
			prefix = ">"
		}
		style := lipgloss.NewStyle().Width(width-4).Padding(0, 1)
		if i == m.selectedSuggestion {
			style = style.Bold(true).Foreground(colorInk).Background(colorBlue)
		}
		rows = append(rows, style.Render(prefix + " " + suggestion))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAmber).Padding(0, 1).Width(width).Render(strings.Join(rows, "\n"))
}

func (m ChatModel) renderProviderForm(width int) string {
	var rows []string
	rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(colorInk).Render("Add Custom Provider"))
	rows = append(rows, "")
	
	// Name field
	nameLabel := "Name:"
	nameStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if m.providerFormFocus == 0 {
		nameStyle = nameStyle.Bold(true).Foreground(colorCyan)
	}
	rows = append(rows, nameLabel + " " + m.providerFormName.View())
	
	// Base URL field
	baseURLLabel := "Base URL:"
	baseURLStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if m.providerFormFocus == 1 {
		baseURLStyle = baseURLStyle.Bold(true).Foreground(colorCyan)
	}
	rows = append(rows, baseURLLabel + " " + m.providerFormBaseURL.View())
	
	// API Key field
	apiKeyLabel := "API Key:"
	apiKeyStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if m.providerFormFocus == 2 {
		apiKeyStyle = apiKeyStyle.Bold(true).Foreground(colorCyan)
	}
	rows = append(rows, apiKeyLabel + " " + m.providerFormAPIKey.View())
	
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(colorSubtle).Render("Tab: switch field  Enter: save  Esc: cancel"))
	
	formContent := strings.Join(rows, "\n")
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(colorCyan).Padding(1, 2).Width(minInt(60, width)).Render(formContent)
}

func (m ChatModel) visibleMessages(maxRows int) []ChatMessage {
	if len(m.messages) == 0 {
		return nil
	}
	maxMessages := maxInt(1, maxRows/4)
	end := len(m.messages) - m.scroll
	end = clampInt(end, 0, len(m.messages))
	start := maxInt(0, end-maxMessages)
	return m.messages[start:end]
}

func (m ChatModel) renderChatFooter(width int) string {
	provider := m.providerName()
	status := fmt.Sprintf("%s · %s", provider, m.status)
	left := lipgloss.NewStyle().Foreground(colorMuted).Render("Enter send  Tab input/sidebar  Ctrl+L clear  PgUp/PgDn scroll  Esc quit")
	right := lipgloss.NewStyle().Foreground(colorSubtle).Render(truncateText(status, width/3))
	space := strings.Repeat(" ", maxInt(1, width-lipgloss.Width(left)-lipgloss.Width(right)))
	return lipgloss.NewStyle().Width(width).Render(left + space + right)
}

func (m ChatModel) helpText(width int) string {
	body := "Ti Chat commands:\n" +
		"  /help              toggle this help\n" +
		"  /providers         show provider list\n" +
		"  /provider <name>   switch provider\n" +
		"  /model             show provider models\n" +
		"  /add-provider      add custom provider\n" +
		"  /auto              show auto commands\n" +
		"  /clear             clear chat history\n" +
		"  /exit              quit\n\n" +
		"Auto commands:\n" +
		"  /auto doctor           - System health check\n" +
		"  /auto bridge           - Cookie + MCP + Tunnel\n" +
		"  /auto snapshot         - Full system backup\n" +
		"  /auto fix-safe         - Safe fix workflow\n" +
		"  /auto pr-review        - Automated PR review\n" +
		"  /auto taskboard-sync   - Sync taskboard\n" +
		"  /auto provider-setup   - Provider configuration\n" +
		"  /auto mcp-publish      - Publish MCP server\n\n" +
		"Shortcuts:\n" +
		"  Tab                switch provider / show suggestions\n" +
		"  Ctrl+L             clear chat\n" +
		"  PgUp/PgDn          scroll messages\n" +
		"  ?                  toggle help"
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorAmber).Foreground(colorMuted).Padding(1, 2).Width(width).Render(body)
}

func (m ChatModel) providerListText() string {
	if len(m.providers) == 0 {
		return "No providers available."
	}
	var lines []string
	for i, name := range m.providers {
		p, _ := m.registry.Get(name)
		model := ""
		healthy := "unknown"
		if p != nil {
			model = p.DefaultModel()
			if p.IsHealthy() {
				healthy = "healthy"
			} else {
				healthy = "configured"
			}
		}
		lines = append(lines, fmt.Sprintf("%d. %s · %s · %s", i+1, name, healthy, model))
	}
	return strings.Join(lines, "\n")
}

func renderMessage(msg ChatMessage, width int) string {
	role := msg.Role
	label := "You"
	color := colorBlue
	symbol := "●"
	if role == "assistant" {
		label = "Ti"
		color = colorPink
		symbol = "◆"
	} else if role == "error" {
		label = "Error"
		color = colorRed
		symbol = "!"
	} else if role == "system" {
		label = "System"
		color = colorMuted
		symbol = "•"
	}
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(color)
	bodyStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(color).Padding(0, 1).Width(maxInt(28, width))
	stamp := msg.Time.Format("15:04")
	body := wrapText(msg.Content, maxInt(24, width-4))
	return labelStyle.Render(fmt.Sprintf("%s %s · %s", symbol, label, stamp)) + "\n" + bodyStyle.Render(body)
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
	out := make([]providers.Message, 0, len(msgs))
	for _, m := range msgs {
		if m.Role != "user" && m.Role != "assistant" && m.Role != "system" {
			continue
		}
		out = append(out, providers.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

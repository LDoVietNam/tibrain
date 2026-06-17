package tmux

// Pane represents a tmux pane
type Pane struct {
	ID       string
	Window   string
	Session  string
	IsActive bool
}

// Window represents a tmux window
type Window struct {
	Name    string
	ID      string
	Session string
	Panes   []Pane
}

// Session represents a tmux session
type Session struct {
	Name    string
	Created string
	Windows []Window
}

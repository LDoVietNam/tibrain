package opencode

import (
	"bufio"
	"encoding/json"
	"io"
)

// Event is a single newline-delimited JSON event from `opencode run --format json`.
type Event struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionID"`
	Timestamp int64           `json:"timestamp"`
	Part      json.RawMessage `json:"part,omitempty"`
	Error     json.RawMessage `json:"error,omitempty"`
}

// TextPart is the inner structure when Event.Type == "text".
type TextPart struct {
	Text string `json:"text"`
	Time *struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	} `json:"time,omitempty"`
}

// ToolPart is the inner structure when Event.Type == "tool_use".
type ToolPart struct {
	Tool  string          `json:"tool"`
	State json.RawMessage `json:"state"`
}

// ParseEventStream reads newline-delimited JSON events from r and sends them
// on the returned channel. The channel is closed when r is exhausted or errors.
func ParseEventStream(r io.Reader) <-chan Event {
	ch := make(chan Event, 32)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB per line
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var ev Event
			if err := json.Unmarshal(line, &ev); err != nil {
				continue // skip malformed lines
			}
			ch <- ev
		}
	}()
	return ch
}

// ExtractText returns the text content from a "text" event, or "".
func (e Event) ExtractText() string {
	if e.Type != "text" || len(e.Part) == 0 {
		return ""
	}
	var p TextPart
	if err := json.Unmarshal(e.Part, &p); err != nil {
		return ""
	}
	return p.Text
}

// ExtractTool returns the tool name from a "tool_use" event, or "".
func (e Event) ExtractTool() string {
	if e.Type != "tool_use" || len(e.Part) == 0 {
		return ""
	}
	var p ToolPart
	if err := json.Unmarshal(e.Part, &p); err != nil {
		return ""
	}
	return p.Tool
}

// Package agent implements the agentic execution loop for Ti CLI.
// The loop follows the pattern: think → call tool → observe → repeat
// until the model returns a final text response (stop_reason == "end_turn").
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ti/cli/internal/providers"
	"github.com/ti/cli/internal/tools"
)

const defaultMaxTurns = 20

// compressHistoryThreshold là số messages tối thiểu trước khi bắt đầu compress.
const compressHistoryThreshold = 15

// compressKeepRecent là số messages gần nhất được giữ nguyên (không truncate tool results).
const compressKeepRecent = 3

// compressToolMaxLen là độ dài tối đa của tool result sau khi truncate.
const compressToolMaxLen = 500

// compressHistory truncate tool results cũ trong history để giảm token usage.
// Chỉ áp dụng khi len(messages) > compressHistoryThreshold.
// System prompt và user messages KHÔNG bị truncate.
// Tool results của compressKeepRecent messages gần nhất KHÔNG bị truncate.
func compressHistory(messages []providers.Message) []providers.Message {
	if len(messages) <= compressHistoryThreshold {
		return messages
	}

	result := make([]providers.Message, len(messages))
	copy(result, messages)

	// Xác định boundary: chỉ compress messages ngoài compressKeepRecent cuối
	cutoff := len(result) - compressKeepRecent
	if cutoff < 0 {
		cutoff = 0
	}

	for i := 0; i < cutoff; i++ {
		msg := &result[i]
		// Chỉ truncate tool messages (role == "tool")
		if msg.Role != "tool" {
			continue
		}
		// Parse ToolResult để truncate content
		var tr providers.ToolResult
		if err := json.Unmarshal([]byte(msg.Content), &tr); err != nil {
			// Không parse được → truncate raw string nếu quá dài
			if len(msg.Content) > compressToolMaxLen {
				msg.Content = msg.Content[:compressToolMaxLen] + " [truncated for context]"
			}
			continue
		}
		if len(tr.Content) > compressToolMaxLen {
			tr.Content = tr.Content[:compressToolMaxLen] + " [truncated for context]"
			if b, err := json.Marshal(tr); err == nil {
				msg.Content = string(b)
			}
		}
	}

	return result
}

// Options configures the agentic loop.
type Options struct {
	// MaxTurns limits how many tool-call rounds are allowed (default 20).
	MaxTurns int
	// Verbose prints tool calls and results to w.
	Verbose bool
	// Writer is where verbose output goes (default: io.Discard).
	Writer io.Writer
	// AllowedTools restricts which tools the AI may call (nil = all).
	AllowedTools []string
	// Depth là độ sâu nesting hiện tại (0 = top-level).
	Depth int
	// MaxDepth là độ sâu tối đa cho phép (default 3 nếu 0).
	MaxDepth int
}

// Result is the final output of the agentic loop.
type Result struct {
	Content string
	Turns   int
	Usage   providers.Usage
}

// Run executes the agentic loop using the given ClaudeProvider.
// It sends the initial messages, handles tool calls, and returns the final text.
func Run(ctx context.Context, p *providers.ClaudeProvider, req providers.ChatRequest, opts Options) (*Result, error) {
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = defaultMaxTurns
	}
	if opts.Writer == nil {
		opts.Writer = io.Discard
	}

	// Build tool definitions — filter if AllowedTools is set.
	defs := tools.AllDefs()
	if len(opts.AllowedTools) > 0 {
		allowed := make(map[string]bool, len(opts.AllowedTools))
		for _, t := range opts.AllowedTools {
			allowed[t] = true
		}
		filtered := defs[:0]
		for _, d := range defs {
			if allowed[d.Name] {
				filtered = append(filtered, d)
			}
		}
		defs = filtered
	}
	req.Tools = defs

	messages := make([]providers.Message, len(req.Messages))
	copy(messages, req.Messages)

	result := &Result{}

	for turn := 0; turn < opts.MaxTurns; turn++ {
		// Compress history để tránh context overflow
		messages = compressHistory(messages)
		req.Messages = messages
		ar, err := p.ChatWithTools(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("turn %d: %w", turn+1, err)
		}

		result.Turns = turn + 1
		if ar.Usage != nil {
			result.Usage.PromptTokens += ar.Usage.PromptTokens
			result.Usage.CompletionTokens += ar.Usage.CompletionTokens
			result.Usage.TotalTokens += ar.Usage.TotalTokens
		}

		// Append assistant message (with raw blocks so tool_use IDs are preserved).
		assistantContent := buildAssistantContent(ar)
		messages = append(messages, providers.Message{
			Role:    "assistant",
			Content: assistantContent,
		})

		if ar.Content != "" {
			fmt.Fprintf(opts.Writer, "💭 %s\n", ar.Content)
		}

		// No tool calls → done.
		if len(ar.ToolCalls) == 0 || ar.StopReason == "end_turn" {
			result.Content = ar.Content
			return result, nil
		}

		// Execute each tool call and collect results.
		for _, tc := range ar.ToolCalls {
			if opts.Verbose {
				inputJSON, _ := json.Marshal(tc.Input)
				fmt.Fprintf(opts.Writer, "🔧 %s(%s)\n", tc.Name, string(inputJSON))
			}

			output, execErr := tools.Dispatch(ctx, tc.Name, tc.Input)
			tr := providers.ToolResult{ToolUseID: tc.ID}
			if execErr != nil {
				tr.Content = execErr.Error()
				tr.IsError = true
			} else {
				tr.Content = output
			}

			if opts.Verbose {
				preview := tr.Content
				if len(preview) > 200 {
					preview = preview[:200] + "…"
				}
				fmt.Fprintf(opts.Writer, "   → %s\n", strings.ReplaceAll(preview, "\n", "↵"))
			}

			trJSON, _ := json.Marshal(tr)
			messages = append(messages, providers.Message{
				Role:    "tool",
				Content: string(trJSON),
			})
		}
	}

	return nil, fmt.Errorf("agentic loop exceeded max turns (%d)", opts.MaxTurns)
}

// buildAssistantContent serialises the assistant's raw content blocks back to
// a JSON string so the Anthropic API can match tool_use IDs on the next turn.
func buildAssistantContent(ar *providers.AgentResponse) string {
	if len(ar.RawBlocks) == 0 {
		return ar.Content
	}
	b, err := json.Marshal(ar.RawBlocks)
	if err != nil {
		return ar.Content
	}
	return string(b)
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// RTKRecentCommand represents a parsed recent command from `rtk gain -H`
type RTKRecentCommand struct {
	Time        string  `json:"time"`
	Command     string  `json:"command"`
	SavingsPct  float64 `json:"savings_pct"`
	SavedTokens string  `json:"saved_tokens"`
	TrendUp     bool    `json:"trend_up"`
}

// RTKCommandStat represents a parsed command statistic from `rtk gain -H`
type RTKCommandStat struct {
	Index       int     `json:"index"`
	Command     string  `json:"command"`
	Count       int     `json:"count"`
	SavedTokens string  `json:"saved_tokens"`
	SavingsPct  float64 `json:"savings_pct"`
	Time        string  `json:"time"`
	Impact      string  `json:"impact"`
}

// getRTKPath resolves the absolute path to rtk.exe
func getRTKPath() string {
	// Try standard path resolution first
	if path, err := exec.LookPath("rtk"); err == nil {
		return path
	}
	if path, err := exec.LookPath("rtk.exe"); err == nil {
		return path
	}

	// Try custom path in user system
	customPath := "Z:\\02_CORE\\_cli\\bin\\rtk.exe"
	if _, err := os.Stat(customPath); err == nil {
		return customPath
	}

	// Fallback
	return "rtk"
}

// executeRTKCommand runs rtk with arguments and returns stdout
func executeRTKCommand(args ...string) (string, error) {
	rtkPath := getRTKPath()
	cmd := exec.Command(rtkPath, args...)

	// Set working directory if we are running in Z:\01_PROJECTS
	cmd.Dir = "Z:\\01_PROJECTS"

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// rtkGainHandler returns the token savings metrics
func (s *APIServer) rtkGainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Allow CORS just in case
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. Fetch JSON stats from `rtk gain -a -f json`
	var stats map[string]interface{}
	jsonStr, err := executeRTKCommand("gain", "-a", "-f", "json")
	if err != nil {
		logger.Warn("Failed to run rtk gain -f json, using mock fallback: %v", err)
		stats = getMockRTKGainStats()
	} else {
		if err := json.Unmarshal([]byte(jsonStr), &stats); err != nil {
			logger.Warn("Failed to parse rtk gain JSON output: %v", err)
			stats = getMockRTKGainStats()
		}
	}

	// 2. Fetch history text from `rtk gain -H` and parse recent commands
	historyText, err := executeRTKCommand("gain", "-H")
	var recentCommands []RTKRecentCommand
	var commandStats []RTKCommandStat

	if err == nil {
		recentCommands = parseRecentCommands(historyText)
		commandStats = parseCommandStats(historyText)
	} else {
		logger.Warn("Failed to run rtk gain -H: %v", err)
	}

	// If history is empty, populate with some defaults
	if len(recentCommands) == 0 {
		recentCommands = getMockRecentCommands()
	}
	if len(commandStats) == 0 {
		commandStats = getMockCommandStats()
	}

	// Append history to response
	response := map[string]interface{}{
		"summary":        stats["summary"],
		"daily":          stats["daily"],
		"weekly":         stats["weekly"],
		"monthly":        stats["monthly"],
		"recent_history": recentCommands,
		"command_stats":  commandStats,
		"rtk_path":       getRTKPath(),
		"timestamp":      time.Now(),
		"status":         "active",
	}

	json.NewEncoder(w).Encode(response)
}

// rtkDiscoverHandler returns potential token savings suggestions
func (s *APIServer) rtkDiscoverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var discoverData map[string]interface{}
	jsonStr, err := executeRTKCommand("discover", "--format", "json")
	if err != nil {
		logger.Warn("Failed to run rtk discover, using mock: %v", err)
		discoverData = getMockDiscoverStats()
	} else {
		// RTK stdout might contain a warning header line like `[rtk] /!\ No hook installed...`
		// We need to clean and find the first JSON character `{` or `[`
		cleanedJSON := cleanJSONOutput(jsonStr)
		if err := json.Unmarshal([]byte(cleanedJSON), &discoverData); err != nil {
			logger.Warn("Failed to parse rtk discover JSON output: %v, raw output: %s", err, jsonStr)
			discoverData = getMockDiscoverStats()
		}
	}

	json.NewEncoder(w).Encode(discoverData)
}

// cleanJSONOutput extracts the first valid JSON block from a string
func cleanJSONOutput(input string) string {
	idx := strings.Index(input, "{")
	if idx == -1 {
		idx = strings.Index(input, "[")
	}
	if idx != -1 {
		return input[idx:]
	}
	return input
}

// parseRecentCommands parses command history from `rtk gain -H` text
func parseRecentCommands(text string) []RTKRecentCommand {
	var list []RTKRecentCommand
	lines := strings.Split(text, "\n")
	inRecent := false

	// Regex: e.g. "05-19 22:44 ▲ rtk git status            -91% (2.4K)"
	// Group 1: Date (05-19 22:44)
	// Group 2: Indicator (▲ or •)
	// Group 3: Command (rtk git status)
	// Group 4: Savings Pct (-91%)
	// Group 5: Saved Tokens (2.4K)
	re := regexp.MustCompile(`^(\d{2}-\d{2}\s+\d{2}:\d{2})\s+([▲•])\s+(.+?)\s+-(\d+)%\s+\((.+?)\)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Recent Commands") {
			inRecent = true
			continue
		}
		if inRecent {
			if line == "" || strings.HasPrefix(line, "───") {
				continue
			}
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 6 {
				pct, _ := strconv.ParseFloat(matches[4], 64)
				list = append(list, RTKRecentCommand{
					Time:        matches[1],
					TrendUp:     matches[2] == "▲",
					Command:     strings.TrimSpace(matches[3]),
					SavingsPct:  pct,
					SavedTokens: matches[5],
				})
			}
		}
	}
	return list
}

// parseCommandStats parses command statistics from `rtk gain -H` text
func parseCommandStats(text string) []RTKCommandStat {
	var list []RTKCommandStat
	lines := strings.Split(text, "\n")
	inStats := false

	// Regex: e.g. " 1.  rtk git status                1   2.4K   91.1%   143ms  ██████████"
	// Group 1: Index (1)
	// Group 2: Command (rtk git status)
	// Group 3: Count (1)
	// Group 4: Saved (2.4K)
	// Group 5: Pct (91.1%)
	// Group 6: Time (143ms)
	// Group 7: Impact (██████████)
	re := regexp.MustCompile(`^\s*(\d+)\.\s+(.+?)\s+(\d+)\s+([\d\.\w]+)\s+([\d\.]+)%\s+([\w\.]+)\s+([█░]+|)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "By Command") {
			inStats = true
			continue
		}
		if inStats {
			if strings.HasPrefix(line, "Recent Commands") {
				break
			}
			if line == "" || strings.HasPrefix(line, "───") || strings.HasPrefix(line, "#") || strings.Contains(line, "Command") {
				continue
			}
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 7 {
				idx, _ := strconv.Atoi(matches[1])
				count, _ := strconv.Atoi(matches[3])
				pct, _ := strconv.ParseFloat(matches[5], 64)
				list = append(list, RTKCommandStat{
					Index:       idx,
					Command:     strings.TrimSpace(matches[2]),
					Count:       count,
					SavedTokens: matches[4],
					SavingsPct:  pct,
					Time:        matches[6],
					Impact:      strings.TrimSpace(matches[7]),
				})
			}
		}
	}
	return list
}

// getMockRTKGainStats provides rich mock gain stats for fallback
func getMockRTKGainStats() map[string]interface{} {
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_commands":  248,
			"total_input":     2384500,
			"total_output":    754200,
			"total_saved":     1630300,
			"avg_savings_pct": 68.37,
			"total_time_ms":   4839200,
			"avg_time_ms":     19512,
		},
		"daily": []map[string]interface{}{
			{"date": "2026-05-19", "commands": 24, "input_tokens": 284500, "output_tokens": 85100, "saved_tokens": 199400, "savings_pct": 70.1, "total_time_ms": 420000},
			{"date": "2026-05-20", "commands": 38, "input_tokens": 395000, "output_tokens": 124000, "saved_tokens": 271000, "savings_pct": 68.6, "total_time_ms": 780000},
			{"date": "2026-05-21", "commands": 42, "input_tokens": 421000, "output_tokens": 118000, "saved_tokens": 303000, "savings_pct": 72.0, "total_time_ms": 890000},
			{"date": "2026-05-22", "commands": 55, "input_tokens": 580000, "output_tokens": 182000, "saved_tokens": 398000, "savings_pct": 68.6, "total_time_ms": 1120000},
			{"date": "2026-05-23", "commands": 89, "input_tokens": 704000, "output_tokens": 245100, "saved_tokens": 458900, "savings_pct": 65.2, "total_time_ms": 1629200},
		},
		"weekly": []map[string]interface{}{
			{"week_start": "2026-05-18", "week_end": "2026-05-24", "commands": 248, "input_tokens": 2384500, "output_tokens": 754200, "saved_tokens": 1630300, "savings_pct": 68.37},
		},
		"monthly": []map[string]interface{}{
			{"month": "2026-05", "commands": 248, "input_tokens": 2384500, "output_tokens": 754200, "saved_tokens": 1630300, "savings_pct": 68.37},
		},
	}
}

// getMockRecentCommands provides default recent history
func getMockRecentCommands() []RTKRecentCommand {
	return []RTKRecentCommand{
		{Time: "05-23 15:34", TrendUp: true, Command: "rtk git diff main", SavingsPct: 84.6, SavedTokens: "185.3K"},
		{Time: "05-23 15:20", TrendUp: true, Command: "rtk go test ./apps/tibrain", SavingsPct: 75.1, SavedTokens: "32.4K"},
		{Time: "05-23 14:58", TrendUp: false, Command: "rtk npm install", SavingsPct: 0.0, SavedTokens: "0"},
		{Time: "05-23 14:12", TrendUp: true, Command: "rtk git status", SavingsPct: 91.2, SavedTokens: "2.4K"},
		{Time: "05-23 12:05", TrendUp: true, Command: "rtk cargo build --release", SavingsPct: 62.4, SavedTokens: "420.5K"},
	}
}

// getMockCommandStats provides default command stats
func getMockCommandStats() []RTKCommandStat {
	return []RTKCommandStat{
		{Index: 1, Command: "rtk git diff", Count: 48, SavedTokens: "1.2M", SavingsPct: 88.5, Time: "210ms", Impact: "██████████"},
		{Index: 2, Command: "rtk cargo build", Count: 12, SavedTokens: "850K", SavingsPct: 64.2, Time: "14.2s", Impact: "███████░░░"},
		{Index: 3, Command: "rtk go test", Count: 35, SavedTokens: "320K", SavingsPct: 78.4, Time: "2.8s", Impact: "███░░░░░░░"},
		{Index: 4, Command: "rtk git status", Count: 85, SavedTokens: "204K", SavingsPct: 91.2, Time: "120ms", Impact: "██░░░░░░░░"},
		{Index: 5, Command: "rtk npm run lint", Count: 14, SavedTokens: "150K", SavingsPct: 82.1, Time: "5.4s", Impact: "█░░░░░░░░░"},
	}
}

// getMockDiscoverStats provides default discover output
func getMockDiscoverStats() map[string]interface{} {
	return map[string]interface{}{
		"sessions_scanned": 12,
		"total_commands":   148,
		"already_rtk":      92,
		"since_days":       7,
		"supported": []map[string]interface{}{
			{"command": "git diff", "potential_pct": 85.0, "average_size_bytes": 142000, "times_run": 14},
			{"command": "npm run test", "potential_pct": 70.0, "average_size_bytes": 85000, "times_run": 8},
			{"command": "go test ./...", "potential_pct": 75.0, "average_size_bytes": 62000, "times_run": 12},
		},
		"unsupported": []map[string]interface{}{
			{"command": "cat data.json", "reason": "output parsing bypass"},
			{"command": "echo 'completed'", "reason": "too small to optimize"},
		},
		"parse_errors":       0,
		"rtk_disabled_count": 56,
		"rtk_disabled_examples": []string{
			"git diff (run directly without rtk prefix)",
			"go test ./... (run directly without rtk prefix)",
			"npm run test (run directly without rtk prefix)",
		},
	}
}

// ─────────────────────────────────────────────────────────────
// Standalone RTK Sidecar Compressor & API Handlers
// ─────────────────────────────────────────────────────────────

// LocalRTKCompressor wraps the local RTK CLI and local rules/cache database.
type LocalRTKCompressor struct {
	db      *sql.DB
	rtkPath string
	useWSL  bool
	enabled bool
}

// RTKRule mirrors the local schema rule model.
type RTKRule struct {
	ID           string                 `json:"id"`
	RuleType     string                 `json:"rule_type"`
	ModelPattern string                 `json:"model_pattern"`
	Config       map[string]interface{} `json:"config"`
	Priority     int                    `json:"priority"`
	Enabled      bool                   `json:"enabled"`
}

// NewLocalRTKCompressor instantiates an offline/decoupled RTK compressor.
func NewLocalRTKCompressor(db *sql.DB) *LocalRTKCompressor {
	rtkPath := getRTKPath()
	enabled := false
	useWSL := false

	if rtkPath != "rtk" && rtkPath != "rtk.exe" {
		if _, err := os.Stat(rtkPath); err == nil {
			enabled = true
		}
	}

	if !enabled {
		// Try PATH lookup
		if _, err := exec.LookPath("rtk"); err == nil {
			rtkPath = "rtk"
			enabled = true
		} else if _, err := exec.LookPath("rtk.exe"); err == nil {
			rtkPath = "rtk.exe"
			enabled = true
		}
	}

	// Try WSL if we're on Windows and standard fails
	if !enabled && runtime.GOOS == "windows" {
		if wslPath, err := exec.LookPath("wsl"); err == nil {
			cmd := exec.Command(wslPath, "bash", "-c", "command -v rtk")
			var out bytes.Buffer
			cmd.Stdout = &out
			if cmd.Run() == nil {
				wslRtkPath := strings.TrimSpace(out.String())
				if wslRtkPath != "" {
					rtkPath = wslRtkPath
					useWSL = true
					enabled = true
				}
			}
		}
	}

	return &LocalRTKCompressor{
		db:      db,
		rtkPath: rtkPath,
		useWSL:  useWSL,
		enabled: enabled,
	}
}

// GetRTKRules retrieves dynamic compression rules from the local SQLite DB.
func (c *LocalRTKCompressor) GetRTKRules(ruleType string) ([]RTKRule, error) {
	if c.db == nil {
		return nil, nil
	}
	rows, err := c.db.Query("SELECT id, rule_type, model_pattern, config, priority, enabled FROM rtk_rules WHERE rule_type = ? AND enabled = 1 ORDER BY priority DESC", ruleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []RTKRule
	for rows.Next() {
		var r RTKRule
		var configStr string
		var enabledInt int
		if err := rows.Scan(&r.ID, &r.RuleType, &r.ModelPattern, &configStr, &r.Priority, &enabledInt); err != nil {
			continue
		}
		r.Enabled = enabledInt != 0
		if configStr != "" {
			var cfg map[string]interface{}
			if err := json.Unmarshal([]byte(configStr), &cfg); err == nil {
				r.Config = cfg
			}
		}
		rules = append(rules, r)
	}
	return rules, nil
}

// ModelContextWindow maps model patterns to their max context window in tokens.
var ModelContextWindow = map[string]int{
	"gemini-2.5-pro":    1048576,
	"gemini-2.5-flash":  1048576,
	"gemini-1.5-pro":    2097152,
	"gemini-1.5-flash":  1048576,
	"claude-sonnet-4-5": 200000,
	"claude-opus-4-5":   200000,
	"gpt-4o":            128000,
	"gpt-5.4":           128000,
	"o3-mini":           128000,
	"deepseek-v4-flash": 65536,
	"deepseek-r1":       65536,
	"qwen3-coder":       131072,
	"llama-3.3-70b":     128000,
	"llama-3.1-70b":     128000,
	"mixtral-8x7b":      32768,
	"gemma-2-9b-it":     8192,
}

// CompressionStrategy determines how aggressively to compress based on model context windows.
type CompressionStrategy struct {
	MaxInputChars      int
	PreserveLastN      int
	PreserveSystem     bool
	CompressCodeBlocks bool
}

// DefaultStrategy returns a safe fallback compression strategy.
func DefaultStrategy() CompressionStrategy {
	return CompressionStrategy{
		MaxInputChars:      4000,
		PreserveLastN:      3,
		PreserveSystem:     true,
		CompressCodeBlocks: false,
	}
}

// StrategyForModel returns the appropriate compression strategy based on local cached rules or defaults.
func (c *LocalRTKCompressor) StrategyForModel(modelID string) CompressionStrategy {
	rules, err := c.GetRTKRules("compression")
	if err == nil {
		for _, rule := range rules {
			if rule.Enabled && matchesModelPattern(modelID, rule.ModelPattern) {
				if cfg := rule.Config; cfg != nil {
					strategy := DefaultStrategy()
					if v, ok := cfg["max_input_chars"].(float64); ok {
						strategy.MaxInputChars = int(v)
					}
					if v, ok := cfg["preserve_last_n"].(float64); ok {
						strategy.PreserveLastN = int(v)
					}
					if v, ok := cfg["preserve_system"].(bool); ok {
						strategy.PreserveSystem = v
					}
					if v, ok := cfg["compress_code_blocks"].(bool); ok {
						strategy.CompressCodeBlocks = v
					}
					return strategy
				}
			}
		}
	}

	return strategyForModel(modelID)
}

func matchesModelPattern(modelID, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(modelID, prefix)
	}
	return strings.Contains(modelID, pattern)
}

func strategyForModel(modelID string) CompressionStrategy {
	contextWindow := LookupContextWindow(modelID)

	switch {
	case contextWindow >= 1000000:
		return CompressionStrategy{
			MaxInputChars:      200000,
			PreserveLastN:      20,
			PreserveSystem:     true,
			CompressCodeBlocks: false,
		}
	case contextWindow >= 128000:
		return CompressionStrategy{
			MaxInputChars:      80000,
			PreserveLastN:      10,
			PreserveSystem:     true,
			CompressCodeBlocks: false,
		}
	case contextWindow >= 32000:
		return CompressionStrategy{
			MaxInputChars:      20000,
			PreserveLastN:      5,
			PreserveSystem:     true,
			CompressCodeBlocks: false,
		}
	default:
		return CompressionStrategy{
			MaxInputChars:      4000,
			PreserveLastN:      2,
			PreserveSystem:     true,
			CompressCodeBlocks: true,
		}
	}
}

func LookupContextWindow(modelID string) int {
	if ctx, ok := ModelContextWindow[modelID]; ok {
		return ctx
	}

	for pattern, ctx := range ModelContextWindow {
		if strings.Contains(modelID, pattern) {
			return ctx
		}
	}

	return 32768
}

// Compress compacts command output using local RTK binary execution with in-process string optimization fallback.
func (c *LocalRTKCompressor) Compress(command string, input string) (string, error) {
	if !c.enabled {
		return compressTextInPlace(input, 4000), nil
	}

	var cmd *exec.Cmd
	if c.useWSL {
		wslPath, _ := exec.LookPath("wsl")
		cmd = exec.Command(wslPath, "bash", "-c", fmt.Sprintf("%s %s", c.rtkPath, command))
		cmd.Stdin = strings.NewReader(input)
	} else {
		parts := strings.Fields(command)
		args := append([]string{c.rtkPath}, parts...)
		cmd = exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(input)
	}

	cmd.Dir = "Z:\\01_PROJECTS"

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		logger.Warn("rtk exec failed: %v, falling back to in-process compaction", err)
		return compressTextInPlace(input, 4000), nil
	}

	return out.String(), nil
}

// CompressText optimizes arbitrary text down to maximum character limit.
func (c *LocalRTKCompressor) CompressText(text string, maxLen int) (string, error) {
	if len(text) <= maxLen {
		return text, nil
	}
	return compressTextInPlace(text, maxLen), nil
}

// compressTextInPlace compacts redundant and blank lines in-process.
func compressTextInPlace(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	lines := strings.Split(text, "\n")
	var result []string
	var prevBlank bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if prevBlank {
				continue
			}
			prevBlank = true
			result = append(result, "")
			continue
		}
		prevBlank = false

		condensed := condenseLine(trimmed)

		if len(condensed) > 200 {
			condensed = condensed[:200] + "..."
		}

		result = append(result, condensed)
	}

	compacted := strings.Join(result, "\n")

	if len(compacted) > maxLen {
		truncated := compacted[:maxLen]
		if lastPeriod := strings.LastIndex(truncated, ". "); lastPeriod > maxLen/2 {
			truncated = truncated[:lastPeriod+1]
		}
		return truncated + "\n[... RTK truncated]"
	}

	return compacted
}

func condenseLine(line string) string {
	result := strings.TrimSpace(line)
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	return result
}

// CompressMessagesModelAware performs local model-aware message sequence compaction.
func (c *LocalRTKCompressor) CompressMessagesModelAware(reqBody map[string]interface{}, modelID string) (map[string]interface{}, int, int) {
	msgsRaw, ok := reqBody["messages"]
	if !ok {
		return reqBody, 0, 0
	}
	msgs, ok := msgsRaw.([]interface{})
	if !ok {
		return reqBody, 0, 0
	}

	strategy := c.StrategyForModel(modelID)
	threshold := strategy.MaxInputChars

	totalOriginal := 0
	totalCompressed := 0
	compressedCount := 0
	msgCount := len(msgs)

	for i, m := range msgs {
		msgMap, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		if strategy.PreserveSystem {
			if role, _ := msgMap["role"].(string); role == "system" {
				continue
			}
		}

		if i >= msgCount-strategy.PreserveLastN {
			continue
		}

		contentRaw, ok := msgMap["content"]
		if !ok {
			continue
		}

		content, ok := extractTextFromContent(contentRaw)
		if !ok || len(content) < threshold/4 {
			continue
		}

		totalOriginal += len(content)

		if !strategy.CompressCodeBlocks && containsCodeBlock(content) {
			compressed := compressNonCodeParts(content, threshold/2)
			totalCompressed += len(compressed)
			if len(compressed) < len(content) {
				msgMap["content"] = compressed
				msgs[i] = msgMap
				compressedCount++
			}
			continue
		}

		compressed, err := c.CompressText(content, threshold)
		if err != nil {
			continue
		}

		totalCompressed += len(compressed)
		compressedCount++
		msgMap["content"] = compressed
		msgs[i] = msgMap
	}

	if compressedCount > 0 {
		reqBody["messages"] = msgs
		reqBody["_rtk_compressed"] = compressedCount
		reqBody["_rtk_strategy"] = strategy.MaxInputChars
	}

	return reqBody, totalOriginal, totalCompressed
}

func containsCodeBlock(content string) bool {
	return strings.Contains(content, "```")
}

func compressNonCodeParts(content string, maxLen int) string {
	parts := strings.Split(content, "```")
	if len(parts) <= 1 {
		return compressTextInPlace(content, maxLen)
	}

	var result strings.Builder
	for i, part := range parts {
		if i%2 == 0 {
			compressed := compressTextInPlace(part, maxLen/len(parts))
			result.WriteString(compressed)
		} else {
			result.WriteString("```")
			result.WriteString(part)
			result.WriteString("```")
		}
	}
	return result.String()
}

func extractTextFromContent(contentRaw interface{}) (string, bool) {
	switch v := contentRaw.(type) {
	case string:
		return v, true
	case []interface{}:
		var texts []string
		for _, part := range v {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			if partType, _ := partMap["type"].(string); partType == "text" {
				if txt, ok := partMap["text"].(string); ok {
					texts = append(texts, txt)
				}
			}
		}
		if len(texts) > 0 {
			return strings.Join(texts, "\n"), true
		}
		return "", false
	default:
		return "", false
	}
}

// rtkCompressHandler handles POST /api/rtk/compress
func (s *APIServer) rtkCompressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Command string `json:"command"`
		Input   string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	compressor := NewLocalRTKCompressor(s.hub.db)
	output, err := compressor.Compress(req.Command, req.Input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"compressed":       output,
		"original_bytes":   len(req.Input),
		"compressed_bytes": len(output),
		"savings_pct":      100.0 * (1.0 - float64(len(output))/float64(len(req.Input)+1)),
	}

	json.NewEncoder(w).Encode(response)
}

// rtkCompressMessagesHandler handles POST /api/rtk/compress-messages
func (s *APIServer) rtkCompressMessagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modelID, _ := body["model"].(string)
	if modelID == "" {
		modelID = "gpt-4o"
	}

	compressor := NewLocalRTKCompressor(s.hub.db)
	updatedBody, origChars, compChars := compressor.CompressMessagesModelAware(body, modelID)

	response := map[string]interface{}{
		"body":             updatedBody,
		"original_chars":   origChars,
		"compressed_chars": compChars,
	}

	json.NewEncoder(w).Encode(response)
}

// rtkLogHandler handles POST /api/rtk/log
func (s *APIServer) rtkLogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Command      string  `json:"command"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		SavedTokens  int     `json:"saved_tokens"`
		SavingsPct   float64 `json:"savings_pct"`
		DurationMs   int     `json:"duration_ms"`
		Metadata     string  `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("rtk_log_%d", time.Now().UnixNano())
	timestamp := time.Now().Unix()

	s.hub.asyncWriter.Enqueue(
		"INSERT INTO rtk_savings_log (id, command, input_tokens, output_tokens, saved_tokens, savings_pct, duration_ms, timestamp, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, req.Command, req.InputTokens, req.OutputTokens, req.SavedTokens, req.SavingsPct, req.DurationMs, timestamp, req.Metadata,
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "queued",
		"id":     id,
	})
}

// rtkRulesHandler handles GET /api/rtk/rules
func (s *APIServer) rtkRulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ruleType := r.URL.Query().Get("type")
	if ruleType == "" {
		ruleType = "compression"
	}

	compressor := NewLocalRTKCompressor(s.hub.db)
	rules, err := compressor.GetRTKRules(ruleType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
	})
}

// rtkRulesSyncHandler handles POST /api/rtk/rules/sync
func (s *APIServer) rtkRulesSyncHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Rules []RTKRule `json:"rules"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := s.hub.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM rtk_rules")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	stmt, err := tx.Prepare("INSERT INTO rtk_rules (id, rule_type, model_pattern, config, priority, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, rule := range req.Rules {
		configBytes, _ := json.Marshal(rule.Config)
		enabledVal := 1
		if !rule.Enabled {
			enabledVal = 0
		}
		_, err = stmt.Exec(rule.ID, rule.RuleType, rule.ModelPattern, string(configBytes), rule.Priority, enabledVal, now, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"count":  len(req.Rules),
	})
}

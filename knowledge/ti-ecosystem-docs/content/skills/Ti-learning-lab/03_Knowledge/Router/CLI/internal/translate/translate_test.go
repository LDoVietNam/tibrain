package translate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanAndConvertClaudeCommand(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := "---\ndescription: Review code\nargument-hint: [target]\nallowed-tools: Read, Grep, Bash\n---\nReview $ARGUMENTS and run npm test if needed."
	if err := os.WriteFile(filepath.Join(dir, ".claude", "commands", "review.md"), []byte(cmd), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := Scan(dir, ScanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings=%d", len(findings))
	}
	pack, _, err := Convert(dir, findings, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Commands) != 1 {
		t.Fatalf("commands=%d", len(pack.Commands))
	}
	if !pack.Commands[0].Sandbox {
		t.Fatalf("expected sandbox risk")
	}
	if !strings.Contains(pack.Commands[0].Prompt, "{{args.arguments}}") {
		t.Fatalf("expected normalized arguments: %s", pack.Commands[0].Prompt)
	}
	y := RenderYAML(pack)
	if !strings.Contains(y, "claude-code") || !strings.Contains(y, "review") || !strings.Contains(y, "risks:") {
		t.Fatalf("unexpected yaml: %s", y)
	}
}

func TestMCPServerConversion(t *testing.T) {
	raw := `{"mcpServers":{"fs":{"command":"npx","args":["@modelcontextprotocol/server-filesystem","."],"env":{"FOO":"BAR"}}}}`
	servers := ParseMCPServers(raw, SourceRef{Format: FormatMCP, Path: "mcp.json", Kind: "mcp_server_config"})
	if len(servers) != 1 || servers[0].Command != "npx" || len(servers[0].Args) != 2 || !servers[0].Sandbox {
		t.Fatalf("unexpected servers: %#v", servers)
	}
}

func TestConflictDetection(t *testing.T) {
	pack := &Pack{Instructions: []Instruction{
		{Content: "Use npm for packages", Source: SourceRef{Path: "AGENTS.md"}},
		{Content: "Use pnpm for packages", Source: SourceRef{Path: "CLAUDE.md"}},
	}}
	conflicts := DetectConflicts(pack.Instructions)
	if len(conflicts) == 0 {
		t.Fatal("expected package-manager conflict")
	}
}

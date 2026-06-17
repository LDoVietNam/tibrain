// Package permission - types and constants for the permission system.
package permission

// PermissionMode constants for easy reference.
const (
	ModeAskStr   string = "ask"
	ModeAutoStr  string = "auto"
	ModeYesStr   string = "yes"
	ModeDenyStr  string = "deny"
	ModeSmartStr string = "smart"
)

// ReadOnlyTools lists tools that only read data.
var ReadOnlyTools = []string{
	"file_read",
	"grep",
	"glob",
	"web_fetch",
	"web_search",
	"agent_read",
	"mcp_read",
	"task_get",
	"task_list",
}

// DangerousTools lists tools that can modify or destroy system state.
var DangerousTools = []string{
	"bash",
	"file_write",
	"file_edit",
	"agent_execute",
	"mcp_execute",
}

// SafeTools lists tools that are generally safe to auto-approve.
var SafeTools = []string{
	"todo_write",
	"task_create",
	"task_update",
	"cost_check",
}

// DestructiveCommands lists bash subcommands that are inherently dangerous.
var DestructiveCommands = []string{
	"rm", "rmdir", "del", "remove-item",
	"mv", "move", "rename",
	"DROP", "DELETE", "TRUNCATE", "ALTER",
	"format", "mkfs", "dd",
	"chmod 777", "icacls /grant",
	"sudo", "su",
	"curl.*|.*bash", "wget.*|.*sh", // Pipe to shell
}

// DefaultDenyPatterns are patterns denied by default for security.
var DefaultDenyPatterns = []string{
	"*:*.env",
	"*:*.key",
	"*:*.pem",
	"*:*.crt",
	"*:.git/*",
	"*:.ssh/*",
	"*:*/secrets/*",
	"*:*/passwords/*",
	"*:*/credentials/*",
	"*:*/.aws/*",
	"*:*/.config/gcloud/*",
}

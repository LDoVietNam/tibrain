package permission

// ResourceType represents the type of resource
type ResourceType string

const (
	ResourceFile      ResourceType = "file"
	ResourceDirectory ResourceType = "directory"
	ResourceNetwork   ResourceType = "network"
	ResourceSystem    ResourceType = "system"
	ResourceConfig    ResourceType = "config"
)

// Action represents an action
type Action string

const (
	ActionRead    Action = "read"
	ActionWrite   Action = "write"
	ActionExecute Action = "execute"
	ActionDelete  Action = "delete"
	ActionCreate  Action = "create"
)

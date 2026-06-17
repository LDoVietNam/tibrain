package verify

import (
	"strings"
	"time"
)

type Profile struct {
	TaskType string        `json:"task_type"`
	Timeout  time.Duration `json:"timeout"`
	Commands []string      `json:"commands,omitempty"`
}

func ProfileForTask(taskType string) Profile {
	taskType = strings.ToLower(strings.TrimSpace(taskType))
	switch taskType {
	case "code", "implement", "review", "refactor", "fix", "debug":
		return Profile{TaskType: taskType, Timeout: 3 * time.Minute, Commands: []string{"go test ./..."}}
	case "plan", "design":
		return Profile{TaskType: taskType, Timeout: 90 * time.Second}
	default:
		return Profile{TaskType: taskType, Timeout: 2 * time.Minute}
	}
}

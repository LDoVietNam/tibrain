package status

// Status represents the current status
type Status string

const (
	StatusIdle      Status = "idle"
	StatusBusy      Status = "busy"
	StatusError     Status = "error"
	StatusCompleted Status = "completed"
	StatusUnknown   Status = "unknown"
)

// DetectionMode represents the detection mode
type DetectionMode string

const (
	DetectionModeSimple   DetectionMode = "simple"
	DetectionModeCombined DetectionMode = "combined"
	DetectionModeAdvanced DetectionMode = "advanced"
)

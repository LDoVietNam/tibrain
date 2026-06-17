package projectmemory

type FileSummary struct {
	Path     string   `json:"path"`
	Kind     string   `json:"kind"`
	Exists   bool     `json:"exists"`
	Bytes    int      `json:"bytes,omitempty"`
	Headings []string `json:"headings,omitempty"`
}

type TaskState struct {
	UpdatedAt string   `json:"updatedAt"`
	Task      string   `json:"task"`
	Status    string   `json:"status"`
	Outputs   []string `json:"outputs"`
	Lesson    string   `json:"lesson"`
}

type CapsuleOptions struct {
	Root string
	Goal string
}

type CapsuleResult struct {
	Root    string   `json:"root"`
	OutDir  string   `json:"outDir"`
	Outputs []string `json:"outputs"`
}

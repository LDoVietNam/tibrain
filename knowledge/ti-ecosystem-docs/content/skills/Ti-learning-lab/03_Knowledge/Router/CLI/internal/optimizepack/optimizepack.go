package optimizepack

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func Checks() []Check {
	var out []Check
	out = append(out, statusCommand("go", "version"))
	out = append(out, statusCommand("git", "--version"))
	out = append(out, statusCommand("docker", "--version"))
	out = append(out, statusCommand("bwrap", "--version"))
	out = append(out, Check{Name: "platform", Status: "ok", Detail: runtime.GOOS + "/" + runtime.GOARCH})
	out = append(out, Check{Name: "GOTOOLCHAIN", Status: "info", Detail: firstNonEmpty(os.Getenv("GOTOOLCHAIN"), "(unset; use local for reproducible Go 1.23 builds)")})
	return out
}

func statusCommand(name string, args ...string) Check {
	p, err := exec.LookPath(name)
	if err != nil {
		return Check{Name: name, Status: "missing", Detail: err.Error()}
	}
	cmd := exec.Command(p, args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return Check{Name: name, Status: "warn", Detail: strings.TrimSpace(string(b))}
	}
	return Check{Name: name, Status: "ok", Detail: strings.TrimSpace(string(b))}
}

func BuildPlan() string {
	return strings.TrimSpace(`Recommended max-optimized release flow:

1. Dependency hygiene
   GOTOOLCHAIN=local GOWORK=off go mod tidy

2. Fast checks
   GOTOOLCHAIN=local GOWORK=off go test ./internal/translate ./internal/blocks ./internal/sandbox

3. Small static binary
   CGO_ENABLED=0 GOTOOLCHAIN=local GOWORK=off go build -trimpath -buildvcs=false -tags "netgo,osusergo" -ldflags "-s -w -buildid=" -o bin/ti-cli ./cmd/ti

4. Optional PGO release
   go build -pgo=default.pgo -trimpath -buildvcs=false -o bin/ti-cli-pgo ./cmd/ti

5. Runtime smoke
   ./bin/ti-cli version
   ./bin/ti-cli doctor
   ./bin/ti-cli sandbox doctor
   ./bin/ti-cli translate scan .
   ./bin/ti-cli block list
`) + "\n"
}

func Summary(checks []Check) string {
	var b strings.Builder
	for _, c := range checks {
		fmt.Fprintf(&b, "%-12s %-8s %s\n", c.Name, c.Status, c.Detail)
	}
	return b.String()
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

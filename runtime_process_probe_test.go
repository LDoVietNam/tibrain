package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRuntimeProcessProbe(t *testing.T) {
	out, err := exec.Command(`C:\Windows\System32\netstat.exe`, "-ano").CombinedOutput()
	if err != nil {
		_ = os.WriteFile(`Z:\02_CORE\_cli\.config\tmp\diagnostic-result.txt`, []byte("netstat-error\n"), 0o600)
		return
	}
	var matches []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":1810") {
			matches = append(matches, strings.TrimSpace(line))
		}
	}
	if len(matches) == 0 {
		matches = append(matches, "no-match")
	}
	_ = os.WriteFile(`Z:\02_CORE\_cli\.config\tmp\diagnostic-result.txt`, []byte(strings.Join(matches, "\n")+"\n"), 0o600)
}
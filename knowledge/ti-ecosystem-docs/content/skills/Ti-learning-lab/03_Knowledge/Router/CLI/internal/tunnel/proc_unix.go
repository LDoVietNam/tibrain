//go:build !windows

package tunnel

import (
	"os/exec"
	"syscall"
)

func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true, // new session, detached from terminal
	}
}

func windowsCmd(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

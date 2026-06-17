//go:build windows

package tunnel

import (
	"os/exec"
	"syscall"
)

func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

// windowsCmd wraps exec.Command with hidden window on Windows.
func windowsCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = detachedSysProcAttr()
	return cmd
}

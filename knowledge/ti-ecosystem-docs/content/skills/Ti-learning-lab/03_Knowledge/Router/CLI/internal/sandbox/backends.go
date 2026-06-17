package sandbox

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type PortableBackend struct{}

func (PortableBackend) Name() string                   { return BackendPortable }
func (PortableBackend) Available(context.Context) bool { return true }
func (PortableBackend) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	req.Backend = BackendPortable
	env := SanitizedEnv(os.Environ(), req.Policy.EnvAllow)
	return runCommand(ctx, req, shellArgv(req.Command), env)
}

type BubblewrapBackend struct{}

func (BubblewrapBackend) Name() string { return BackendBubblewrap }
func (BubblewrapBackend) Available(ctx context.Context) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("bwrap")
	return err == nil
}
func (BubblewrapBackend) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	req.Backend = BackendBubblewrap
	args := []string{"--unshare-all", "--die-with-parent"}
	if !req.AllowNetwork {
		args = append(args, "--unshare-net")
	}
	for _, p := range []string{"/usr", "/bin", "/lib", "/lib64", "/etc/resolv.conf"} {
		if _, err := os.Stat(p); err == nil {
			args = append(args, "--ro-bind", p, p)
		}
	}
	args = append(args, "--dev", "/dev", "--proc", "/proc", "--tmpfs", "/tmp")
	cwd := req.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	cwd, _ = filepath.Abs(cwd)
	args = append(args, "--bind", cwd, cwd, "--chdir", cwd)
	args = append(args, "--setenv", "HOME", cwd)
	args = append(args, "--", "sh", "-c", req.Command)
	env := SanitizedEnv(os.Environ(), req.Policy.EnvAllow)
	return runCommand(ctx, req, append([]string{"bwrap"}, args...), env)
}

type DockerBackend struct{}

func (DockerBackend) Name() string { return BackendDocker }
func (DockerBackend) Available(ctx context.Context) bool {
	_, err := exec.LookPath("docker")
	return err == nil
}
func (DockerBackend) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	req.Backend = BackendDocker
	cwd := req.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	cwd, _ = filepath.Abs(cwd)
	image := os.Getenv("TI_SANDBOX_DOCKER_IMAGE")
	if image == "" {
		image = "golang:1.23-bookworm"
	}
	args := []string{"run", "--rm", "-v", cwd + ":/workspace", "-w", "/workspace"}
	if !req.AllowNetwork {
		args = append(args, "--network", "none")
	}
	args = append(args, image, "sh", "-c", req.Command)
	env := SanitizedEnv(os.Environ(), req.Policy.EnvAllow)
	return runCommand(ctx, req, append([]string{"docker"}, args...), env)
}

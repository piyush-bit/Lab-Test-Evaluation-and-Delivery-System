//go:build docker_integration
// +build docker_integration

package docker

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCommandWithRealDocker(t *testing.T) {
	const image = "alpine:3.20"

	binary := requireDocker(t, os.Getenv("DOCKER_BINARY"))
	hostDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(hostDir, "input.txt"), []byte("from-host\n"), 0644); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := RunConfig{
		ImageConfig: ImageConfig{
			Image:  image,
			Binary: binary,
			Stdout: &stdout,
			Stderr: &stderr,
		},
		HostPath:  hostDir,
		Workdir:   "/workspace",
		Command:   `test "$(pwd)" = "/workspace" && test "$(cat input.txt)" = "from-host" && printf 'from-container\n' > output.txt`,
		Timeout:   45 * time.Second,
		MemoryMB:  64,
		PidsLimit: 64,
	}.RunCommand()
	if err != nil {
		t.Fatalf("RunCommand failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	output, err := os.ReadFile(filepath.Join(hostDir, "output.txt"))
	if err != nil {
		t.Fatalf("read container output file: %v", err)
	}
	if string(output) != "from-container\n" {
		t.Fatalf("unexpected container output %q", output)
	}
}

func requireDocker(t *testing.T, binaryName string) string {
	t.Helper()

	binary, err := ImageConfig{Binary: binaryName}.dockerBinary()
	if err != nil {
		t.Skip(err)
	}

	command := exec.Command(binary, "info")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Skipf("docker daemon is not available: %v: %s", err, strings.TrimSpace(stderr.String()))
	}

	return binary
}

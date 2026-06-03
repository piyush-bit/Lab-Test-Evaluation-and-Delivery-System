package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBinary  = "docker"
	defaultWorkdir = "/exercise"
	defaultTimeout = 30 * time.Second
)

//
// Base config (shared by image + run)
//
type ImageConfig struct {
	Image   string
	Binary  string
	Stdout  io.Writer
	Stderr  io.Writer
}

//
// Run config extends ImageConfig via embedding
//
type RunConfig struct {
	ImageConfig

	HostPath  string
	Workdir   string
	Command   string
	Timeout   time.Duration
	MemoryMB  int
	PidsLimit int
}

//
// Shared normalization helpers
//
func (c ImageConfig) dockerBinary() (string, error) {
	binary := strings.TrimSpace(c.Binary)
	if binary == "" {
		binary = defaultBinary
	}

	if _, err := exec.LookPath(binary); err != nil {
		return "", fmt.Errorf("%s is not installed: %w", binary, err)
	}

	return binary, nil
}

func (c ImageConfig) output() (io.Writer, io.Writer) {
	stdout := c.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	stderr := c.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	return stdout, stderr
}
//
// Works for both ImageConfig and RunConfig
//
func (config ImageConfig) EnsureImage() error {
	base := config

	image := strings.TrimSpace(base.Image)
	if image == "" {
		return fmt.Errorf("docker image is required")
	}

	binary, err := base.dockerBinary()
	if err != nil {
		return err
	}

	stdout, stderr := base.output()

	inspect := exec.Command(binary, "image", "inspect", image)
	inspect.Stdout = io.Discard
	inspect.Stderr = io.Discard

	if err := inspect.Run(); err == nil {
		return nil
	}

	pull := exec.Command(binary, "pull", image)
	pull.Stdout = stdout
	pull.Stderr = stderr

	if err := pull.Run(); err != nil {
		return fmt.Errorf("prepare docker image %q: %w", image, err)
	}

	return nil
}

//
// Main run command
//
func (config RunConfig)RunCommand() error {
	if err := config.EnsureImage(); err != nil {
		return err
	}

	binary, err := config.dockerBinary()
	if err != nil {
		return err
	}

	hostPath, err := filepath.Abs(config.HostPath)
	if err != nil {
		return fmt.Errorf("resolve docker host path: %w", err)
	}

	workdir := strings.TrimSpace(config.Workdir)
	if workdir == "" {
		workdir = defaultWorkdir
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	stdout, stderr := config.output()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{
		"run",
		"--rm",
		"--workdir", workdir,
		"--volume", hostPath + ":" + workdir,
	}

	if config.MemoryMB > 0 {
		args = append(args, "--memory", fmt.Sprintf("%dm", config.MemoryMB))
	}

	if config.PidsLimit > 0 {
		args = append(args, "--pids-limit", strconv.Itoa(config.PidsLimit))
	}

	args = append(args,
		config.Image,
		"/bin/sh", "-c", config.Command,
	)

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return context.DeadlineExceeded
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 124 {
		return context.DeadlineExceeded
	}

	if err != nil {
		return fmt.Errorf("docker run failed: %w", err)
	}

	return nil
}
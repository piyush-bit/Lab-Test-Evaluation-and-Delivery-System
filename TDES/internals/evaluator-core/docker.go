package evaluatorcore

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

const (
	defaultBinary         = "docker"
	defaultWorkdir        = "/exercise"
	defaultCommandTimeout = 30 * time.Second
	defaultContainerCmd   = "while true; do sleep 3600; done"
)

type Runtime interface {
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, imageConfig ImageConfig, containerID string) error
	ExecCommand(config ExecConfig) error
	RemoveContainer(ctx context.Context, imageConfig ImageConfig, containerID string) error
}

type ImageConfig struct {
	Image  string
	Binary string
	Stdout io.Writer
	Stderr io.Writer
}

type ContainerConfig struct {
	ImageConfig

	HostPath  string
	Workdir   string
	MemoryMB  int
	PidsLimit int
}

type ExecConfig struct {
	ImageConfig

	ContainerID string
	Command     string
	Timeout     time.Duration
}

type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.Code)
}

type DockerRuntime struct{}

func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{}
}

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

func (c ImageConfig) dockerHost() (string, error) {
	binary := strings.TrimSpace(c.Binary)
	switch {
	case binary == "", binary == defaultBinary, filepath.Base(binary) == defaultBinary:
		return "", nil
	case strings.Contains(binary, "://"):
		return binary, nil
	default:
		return "", fmt.Errorf("docker sdk runtime does not support custom docker binary %q", binary)
	}
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

func (c ImageConfig) newClient() (*client.Client, error) {
	host, err := c.dockerHost()
	if err != nil {
		return nil, err
	}

	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return cli, nil
}

func (c ImageConfig) EnsureImage() error {
	imageName := strings.TrimSpace(c.Image)
	if imageName == "" {
		return fmt.Errorf("docker image is required")
	}

	cli, err := c.newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	stdout, stderr := c.output()

	if _, err := cli.ImageInspect(context.Background(), imageName); err == nil {
		return nil
	}

	reader, err := cli.ImagePull(context.Background(), imageName, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("prepare docker image %q: %w", imageName, err)
	}
	defer reader.Close()

	if _, err := io.Copy(stdout, reader); err != nil {
		return fmt.Errorf("stream docker image pull for %q: %w", imageName, err)
	}

	_ = stderr

	return nil
}

func (c *DockerRuntime) CreateContainer(ctx context.Context, config ContainerConfig) (string, error) {
	if err := config.EnsureImage(); err != nil {
		return "", err
	}

	cli, err := config.newClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	hostPath, err := filepath.Abs(config.HostPath)
	if err != nil {
		return "", fmt.Errorf("resolve docker host path: %w", err)
	}

	workdir := strings.TrimSpace(config.Workdir)
	if workdir == "" {
		workdir = defaultWorkdir
	}

	resources := container.Resources{}
	if config.MemoryMB > 0 {
		resources.Memory = int64(config.MemoryMB) * 1024 * 1024
	}
	if config.PidsLimit > 0 {
		pidsLimit := int64(config.PidsLimit)
		resources.PidsLimit = &pidsLimit
	}

	resp, err := cli.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config: &container.Config{
				Image:      config.Image,
				WorkingDir: workdir,
				Cmd:        []string{"/bin/sh", "-c", defaultContainerCmd},
			},
			HostConfig: &container.HostConfig{
				Binds:     []string{hostPath + ":" + workdir},
				Resources: resources,
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("docker create failed: %w", err)
	}

	containerID := strings.TrimSpace(resp.ID)
	if containerID == "" {
		return "", fmt.Errorf("docker create returned an empty container id")
	}
	return containerID, nil
}

func (c *DockerRuntime) StartContainer(ctx context.Context, imageConfig ImageConfig, containerID string) error {
	cli, err := imageConfig.newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if _, err := cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("docker start failed: %w", err)
	}
	return nil
}

func (c *DockerRuntime) ExecCommand(config ExecConfig) error {
	cli, err := config.newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}

	stdout, stderr := config.output()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := cli.ExecCreate(ctx, config.ContainerID, client.ExecCreateOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/sh", "-c", config.Command},
	})
	if err != nil {
		return fmt.Errorf("docker exec create failed: %w", err)
	}

	attach, err := cli.ExecAttach(ctx, resp.ID, client.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("docker exec attach failed: %w", err)
	}
	defer attach.Close()

	_, copyErr := stdcopy.StdCopy(stdout, stderr, attach.Reader)

	if ctx.Err() == context.DeadlineExceeded {
		return context.DeadlineExceeded
	}
	if copyErr != nil {
		return fmt.Errorf("docker exec failed: %w", copyErr)
	}

	inspect, err := cli.ExecInspect(context.Background(), resp.ID, client.ExecInspectOptions{})
	if err != nil {
		return fmt.Errorf("docker exec inspect failed: %w", err)
	}
	if inspect.ExitCode == 0 {
		return nil
	}

	return &ExitError{Code: inspect.ExitCode}
}

func (c *DockerRuntime) RemoveContainer(ctx context.Context, imageConfig ImageConfig, containerID string) error {
	cli, err := imageConfig.newClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if _, err := cli.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("docker rm failed: %w", err)
	}
	return nil
}

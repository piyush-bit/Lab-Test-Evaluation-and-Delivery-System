package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"TDES/internals/docker"
	"TDES/internals/exercise"
)

type Config struct {
	ExercisePath string
	DockerBinary string
	Stdout       io.Writer
	Stderr       io.Writer
}

func RunTestsDocker(config Config) error {
	exercisePath, err := resolveExercisePath(config.ExercisePath)
	if err != nil {
		return err
	}

	manifest, err := exercise.LoadManifest(exercisePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(manifest.RunnerImage) == "" {
		return fmt.Errorf("manifest is missing runner_image")
	}
	if strings.TrimSpace(manifest.LocalEntrypoint) == "" {
		return fmt.Errorf("manifest is missing local_entrypoint")
	}

	timeout := time.Duration(manifest.Limits.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	err = docker.RunConfig{
		ImageConfig: docker.ImageConfig{
			Image:  manifest.RunnerImage,
			Binary: config.DockerBinary,
			Stdout: config.Stdout,
			Stderr: config.Stderr,
		},
		HostPath:  exercisePath,
		Command:   manifest.LocalEntrypoint,
		Timeout:   timeout,
		MemoryMB:  manifest.Limits.MemoryMB,
		PidsLimit: manifest.Limits.PidsLimit,
	}.RunCommand()
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("local tests timed out after %s", timeout)
	}
	if err != nil {
		return fmt.Errorf("local tests failed: %w", err)
	}
	return nil
}

func RunTestsLocal(config Config) error {
	exercisePath, err := resolveExercisePath(config.ExercisePath)
	if err != nil {
		return err
	}

	manifest, err := exercise.LoadManifest(exercisePath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(manifest.LocalEntrypoint) == "" {
		return fmt.Errorf("manifest is missing local_entrypoint")
	}

	stdout := config.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := config.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	timeout := time.Duration(manifest.Limits.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	command := exec.CommandContext(ctx, "/bin/sh", "-c", manifest.LocalEntrypoint)
	command.Dir = exercisePath
	command.Stdout = stdout
	command.Stderr = stderr

	err = command.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("local tests timed out after %s", timeout)
	}
	if err != nil {
		return fmt.Errorf("local tests failed: %w", err)
	}
	return nil
}

func ExecDocker(config Config) error {
	return RunTestsDocker(config)
}

func resolveExercisePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve exercise path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("access exercise path %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("exercise path %q is not a directory", absPath)
	}

	if _, err := os.Stat(filepath.Join(absPath, "manifest.json")); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("manifest.json not found in %q", absPath)
		}
		return "", fmt.Errorf("access manifest.json: %w", err)
	}

	return absPath, nil
}

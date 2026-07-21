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

type GradingTestResult struct {
	Command        string `json:"command"`
	PointsPossible int    `json:"points_possible"`
	PointsEarned   int    `json:"points_earned"`
	Status         string `json:"status"` // "pass" | "fail"
	Output         string `json:"output"`
	Public         bool   `json:"public"`
}

type GradingResult struct {
	Success      bool                `json:"success"`
	EarnedPoints int                 `json:"earned_points"`
	MaxPoints    int                 `json:"max_points"`
	Results      []GradingTestResult `json:"results"`
}

func RunGradingTests(config Config, useLocal bool, filterCommand string) (GradingResult, error) {
	exercisePath, err := resolveExercisePath(config.ExercisePath)
	if err != nil {
		return GradingResult{}, err
	}

	manifest, err := exercise.LoadManifest(exercisePath)
	if err != nil {
		return GradingResult{}, err
	}

	timeout := time.Duration(manifest.Limits.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	maxPoints := 0
	for _, entry := range manifest.Grading {
		maxPoints += entry.Points
	}

	var results []GradingTestResult
	earnedPoints := 0
	overallSuccess := true

	for _, entry := range manifest.Grading {
		// Filter by specific command if requested
		if filterCommand != "" && entry.Command != filterCommand {
			// Skip and mark as idle
			results = append(results, GradingTestResult{
				Command:        entry.Command,
				PointsPossible: entry.Points,
				Status:         "idle",
				Public:         entry.Public,
				Output:         "Skipped.",
			})
			continue
		}

		// Private tests are locked locally in student workspace
		if !entry.Public {
			results = append(results, GradingTestResult{
				Command:        entry.Command,
				PointsPossible: entry.Points,
				Status:         "locked",
				Public:         entry.Public,
				Output:         "This is a private test case. It is locked and will be executed upon submission.",
			})
			continue
		}

		var outputBuf strings.Builder
		testResult := GradingTestResult{
			Command:        entry.Command,
			PointsPossible: entry.Points,
			Status:         "fail",
			Public:         entry.Public,
		}

		if useLocal {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			command := exec.CommandContext(ctx, "/bin/sh", "-c", entry.Command)
			command.Dir = exercisePath
			command.Stdout = &outputBuf
			command.Stderr = &outputBuf

			err = command.Run()
			cancel()
		} else {
			err = docker.RunConfig{
				ImageConfig: docker.ImageConfig{
					Image:  manifest.RunnerImage,
					Binary: config.DockerBinary,
					Stdout: &outputBuf,
					Stderr: &outputBuf,
				},
				HostPath:  exercisePath,
				Command:   entry.Command,
				Timeout:   timeout,
				MemoryMB:  manifest.Limits.MemoryMB,
				PidsLimit: manifest.Limits.PidsLimit,
			}.RunCommand()
		}

		testResult.Output = outputBuf.String()
		if err != nil {
			overallSuccess = false
			if strings.TrimSpace(testResult.Output) == "" {
				testResult.Output = err.Error()
			}
		} else {
			testResult.Status = "pass"
			testResult.PointsEarned = entry.Points
			earnedPoints += entry.Points
		}
		results = append(results, testResult)
	}

	if results == nil {
		results = []GradingTestResult{}
	}

	return GradingResult{
		Success:      overallSuccess,
		EarnedPoints: earnedPoints,
		MaxPoints:    maxPoints,
		Results:      results,
	}, nil
}

package exercise

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"TDES/internals/docker"
)

// PackageType indicates whether a built package is intended for public (student-facing) or private (instructor/evaluator) distribution.
type PackageType string

const (
	PackageTypePublic  PackageType = "public"
	PackageTypePrivate PackageType = "private"
)

const (
	manifestFileName          = "manifest.json"
	publicPackageArchiveName  = "public-exercise.tar.gz"
	privatePackageArchiveName = "private-exercise.tar.gz"
	publicEntrypoint          = "make test-public"
)

type Exercise struct {
	Manifest *ExerciseManifest
	Path     string
}

func GenerateExerciseTemplate(path string) error {
	// TODO: Implement
	return fmt.Errorf("Not implemented")
}

// TestExercise validates that an exercise runs successfully in its test
// environment.
func (e *Exercise) TestExercise() error {
	if e == nil {
		return fmt.Errorf("exercise is nil")
	}
	if e.Path == "" {
		return fmt.Errorf("exercise path is required")
	}

	if e.Manifest == nil {
		manifest, err := LoadManifest(e.Path)
		if err != nil {
			return err
		}
		e.Manifest = manifest
	}

	if err := e.Manifest.ValidateExerciseContract(); err != nil {
		return err
	}

	// 1. Back up original files
	backups := make(map[string][]byte)
	for _, relPath := range e.Manifest.Submission.IncludePaths {
		livePath := filepath.Join(e.Path, relPath)
		data, err := os.ReadFile(livePath)
		if err == nil {
			backups[relPath] = data
		}
	}

	// Restore backups on completion
	defer func() {
		for relPath, data := range backups {
			livePath := filepath.Join(e.Path, relPath)
			_ = os.WriteFile(livePath, data, 0644)
		}
	}()

	// 2. Copy reference files to live paths so tests run against the solved reference package
	for _, relPath := range e.Manifest.Submission.IncludePaths {
		refPath := filepath.Join(e.Path, "reference", relPath)
		livePath := filepath.Join(e.Path, relPath)
		refData, err := os.ReadFile(refPath)
		if err != nil {
			return fmt.Errorf("failed to read reference file %s: %w", refPath, err)
		}
		err = os.WriteFile(livePath, refData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write live file %s: %w", livePath, err)
		}
	}

	timeout := time.Duration(e.Manifest.Limits.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	for _, entry := range e.Manifest.Grading {
		command := strings.TrimSpace(entry.Command)
		err := docker.RunConfig{
			ImageConfig: docker.ImageConfig{
				Image: e.Manifest.RunnerImage,
			},
			HostPath:  e.Path,
			Command:   command,
			Timeout:   timeout,
			MemoryMB:  e.Manifest.Limits.MemoryMB,
			PidsLimit: e.Manifest.Limits.PidsLimit,
		}.RunCommand()
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("grading command %q timed out after %s", command, timeout)
		}
		if err != nil {
			return fmt.Errorf("grading command %q failed: %w", command, err)
		}
	}

	return nil
}

package exercise

import (
	"context"
	"errors"
	"fmt"
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

package exercise

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ExerciseManifest struct {
	LabID           string         `json:"lab_id"`
	Title           string         `json:"title"`
	Version         string         `json:"version"`
	Language        string         `json:"language"`
	RunnerImage     string         `json:"runner_image"`
	LocalEntrypoint string         `json:"local_entrypoint"`
	Grading         []GradingEntry `json:"grading"`
	Submission      Submission     `json:"submission"`
	Limits          Limits         `json:"limits"`
}

type GradingEntry struct {
	Command string `json:"command"`
	Points  int    `json:"points"`
}

type Submission struct {
	IncludePaths []string `json:"include_paths"`
	PrivateGlobs []string `json:"private_globs"`
	ExcludeGlobs []string `json:"exclude_globs"`
}

type Limits struct {
	MemoryMB       int `json:"memory_mb"`
	TimeoutSeconds int `json:"timeout_seconds"`
	PidsLimit      int `json:"pids_limit"`
}

// LoadManifest returns the exercise manifest from an exercise directory.
func LoadManifest(exercisePath string) (*ExerciseManifest, error) {
	manifestPath := filepath.Join(exercisePath, manifestFileName)
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()

	var manifest ExerciseManifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}

	return &manifest, nil
}

func (m *ExerciseManifest) ValidateExerciseContract() error {
	if m == nil {
		return fmt.Errorf("manifest is required")
	}
	// why ?
	if strings.TrimSpace(m.LocalEntrypoint) != publicEntrypoint {
		return fmt.Errorf("manifest local_entrypoint must be %q", publicEntrypoint)
	}
	if strings.TrimSpace(m.RunnerImage) == "" {
		return fmt.Errorf("manifest is missing runner_image")
	}
	if len(m.Grading) == 0 {
		return fmt.Errorf("manifest is missing grading")
	}
	for i, entry := range m.Grading {
		if strings.TrimSpace(entry.Command) == "" {
			return fmt.Errorf("manifest grading[%d].command is required", i)
		}
		if entry.Points <= 0 {
			return fmt.Errorf("manifest grading[%d].points must be greater than zero", i)
		}
	}
	if len(m.Submission.IncludePaths) == 0 {
		return fmt.Errorf("manifest is missing submission.include_paths")
	}
	for i, includePath := range m.Submission.IncludePaths {
		if strings.TrimSpace(includePath) == "" {
			return fmt.Errorf("manifest submission.include_paths[%d] is required", i)
		}
	}
	for i, privateGlob := range m.Submission.PrivateGlobs {
		if strings.TrimSpace(privateGlob) == "" {
			return fmt.Errorf("manifest submission.private_globs[%d] is required", i)
		}
	}
	for i, excludeGlob := range m.Submission.ExcludeGlobs {
		if strings.TrimSpace(excludeGlob) == "" {
			return fmt.Errorf("manifest submission.exclude_globs[%d] is required", i)
		}
	}
	return nil
}

package exercise

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExerciseRequiresExercisePath(t *testing.T) {
	exercise := &Exercise{}

	err := exercise.TestExercise()
	if err == nil {
		t.Fatal("expected missing path error")
	}
	if !strings.Contains(err.Error(), "exercise path is required") {
		t.Fatalf("expected exercise path error, got %v", err)
	}
}

func TestExerciseLoadsManifestAndValidatesContract(t *testing.T) {
	exerciseDir := t.TempDir()
	writeManifestForTest(t, exerciseDir, validManifestForTest())

	exercise := &Exercise{Path: exerciseDir}
	err := exercise.TestExercise()
	if err == nil {
		t.Fatal("expected manifest validation error")
	}
	if exercise.Manifest == nil {
		t.Fatal("expected TestExercise to load manifest into exercise")
	}
	if exercise.Manifest.LabID != "go101-lab01" {
		t.Fatalf("expected loaded manifest lab id, got %q", exercise.Manifest.LabID)
	}
}

func validManifestForTest() ExerciseManifest {
	return ExerciseManifest{
		LabID:           "go101-lab01",
		RunnerImage:     "golang:1.22",
		LocalEntrypoint: publicEntrypoint,
		Grading: []GradingEntry{
			{Command: "", Points: 100},
		},
		Submission: Submission{
			IncludePaths: []string{"stack.go"},
		},
	}
}

func writeManifestForTest(t *testing.T, dir string, manifest ExerciseManifest) {
	t.Helper()

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, manifestFileName), data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

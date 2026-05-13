package run

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"euc2/internals/exercise"
)

func TestRunTestsLocalRunsEntrypointInExerciseDirectory(t *testing.T) {
	exerciseDir := t.TempDir()
	writeManifest(t, exerciseDir, exercise.ExerciseManifest{
		LocalEntrypoint: "printf '%s' \"$PWD\" > pwd.txt",
		Limits: exercise.Limits{
			TimeoutSeconds: 5,
		},
	})

	if err := RunTestsLocal(Config{ExercisePath: exerciseDir}); err != nil {
		t.Fatalf("RunTestsLocal failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(exerciseDir, "pwd.txt"))
	if err != nil {
		t.Fatalf("read pwd.txt: %v", err)
	}
	if string(data) != exerciseDir {
		t.Fatalf("expected command to run in %q, got %q", exerciseDir, string(data))
	}
}

func writeManifest(t *testing.T, dir string, manifest exercise.ExerciseManifest) {
	t.Helper()

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

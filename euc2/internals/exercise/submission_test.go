package exercise

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSubmissionPackage(t *testing.T) {
	exerciseDir := t.TempDir()
	writeManifestForTest(t, exerciseDir, ExerciseManifest{
		LabID:   "go101-lab01",
		Version: "1.0.0",
		Submission: Submission{
			IncludePaths: []string{"stack.go"},
		},
	})
	writeFileForTest(t, filepath.Join(exerciseDir, "stack.go"), "package main")

	archiveFile, err := os.CreateTemp(t.TempDir(), "*.tar")
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer archiveFile.Close()

	exercise := &Exercise{Path: exerciseDir}
	if err := exercise.CreateSubmissionPackage(archiveFile, "org-1", "student-42"); err != nil {
		t.Fatalf("CreateSubmissionPackage failed: %v", err)
	}

	manifest, err := ReadSubmissionArchiveManifest(archiveFile.Name())
	if err != nil {
		t.Fatalf("ReadSubmissionArchiveManifest failed: %v", err)
	}
	if manifest.OrgID != "org-1" {
		t.Fatalf("expected org id org-1, got %q", manifest.OrgID)
	}
	if manifest.StudentID != "student-42" {
		t.Fatalf("expected student id student-42, got %q", manifest.StudentID)
	}
	if manifest.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %q", manifest.Version)
	}
	if len(manifest.IncludedPaths) != 1 || manifest.IncludedPaths[0] != "stack.go" {
		t.Fatalf("expected included path stack.go, got %v", manifest.IncludedPaths)
	}
}

func writeFileForTest(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

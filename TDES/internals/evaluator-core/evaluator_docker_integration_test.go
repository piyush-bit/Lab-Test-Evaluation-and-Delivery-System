//go:build docker_integration
// +build docker_integration

package evaluatorcore

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"TDES/internals/exercise"
	exerciseinit "TDES/internals/init"
)

func TestEvaluateSubmissionWithRealDocker(t *testing.T) {
	repoRoot := requireRepoRoot(t)
	demoExerciseRoot := filepath.Join(repoRoot, "demo_exercises", "go101-lab01-stack")
	referenceFile := filepath.Join(demoExerciseRoot, "reference", "stack.go")

	binary := requireDocker(t, os.Getenv("DOCKER_BINARY"))
	requireDockerImage(t, binary, "lab-go-runner:v1.0")

	publicPackagePath, privatePackagePath, err := exercise.PackageExercise(demoExerciseRoot)
	if err != nil {
		t.Fatalf("package exercise: %v", err)
	}
	defer os.Remove(publicPackagePath)
	defer os.Remove(privatePackagePath)

	initDir := filepath.Join(t.TempDir(), "workspace")
	if err := exerciseinit.Init(publicPackagePath, initDir); err != nil {
		t.Fatalf("init exercise: %v", err)
	}

	untouchedSubmission := filepath.Join(t.TempDir(), "submission-untouched.tar")
	createSubmissionPackage(t, initDir, untouchedSubmission)

	referenceBytes, err := os.ReadFile(referenceFile)
	if err != nil {
		t.Fatalf("read reference file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(initDir, "stack.go"), referenceBytes, 0644); err != nil {
		t.Fatalf("copy reference into workspace: %v", err)
	}

	referenceSubmission := filepath.Join(t.TempDir(), "submission-reference.tar")
	createSubmissionPackage(t, initDir, referenceSubmission)

	testCases := []struct {
		name         string
		submission   string
		wantEarned   int
		wantStatuses []string
	}{
		{
			name:         "untouched submission",
			submission:   untouchedSubmission,
			wantEarned:   2,
			wantStatuses: []string{"fail", "fail", "fail", "pass", "fail"},
		},
		{
			name:         "reference submission",
			submission:   referenceSubmission,
			wantEarned:   10,
			wantStatuses: []string{"pass", "pass", "pass", "pass", "pass"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluator, err := NewEvaluator(&localArtifactProvider{path: privatePackagePath}, NewDockerRuntime())
			if err != nil {
				t.Fatalf("new evaluator: %v", err)
			}

			result, err := evaluator.EvaluateSubmission(context.Background(), EvaluationRequest{
				SubmissionArchivePath: tc.submission,
				DockerBinary:          binary,
			})
			if err != nil {
				t.Fatalf("EvaluateSubmission failed: %v", err)
			}

			if result.Status != "completed" {
				t.Fatalf("expected completed, got %q", result.Status)
			}
			if result.EarnedPoints != tc.wantEarned || result.MaxPoints != 10 {
				t.Fatalf("unexpected score: %+v", result)
			}

			gotStatuses := make([]string, 0, len(result.Results))
			for _, testResult := range result.Results {
				gotStatuses = append(gotStatuses, testResult.Status)
			}
			if strings.Join(gotStatuses, ",") != strings.Join(tc.wantStatuses, ",") {
				t.Fatalf("unexpected statuses: got=%v want=%v", gotStatuses, tc.wantStatuses)
			}
		})
	}
}

func requireRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for root := filepath.Clean(wd); ; root = filepath.Dir(root) {
		if _, err := os.Stat(filepath.Join(root, "demo_exercises", "go101-lab01-stack")); err == nil {
			if _, err := os.Stat(filepath.Join(root, "TDES")); err == nil {
				return root
			}
		}

		parent := filepath.Dir(root)
		if parent == root {
			break
		}
	}

	t.Fatalf("locate repo root from %s", wd)
	return ""
}

func requireDocker(t *testing.T, binaryName string) string {
	t.Helper()

	binary, err := ImageConfig{Binary: binaryName}.dockerBinary()
	if err != nil {
		t.Skip(err)
	}

	command := exec.Command(binary, "info")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Skipf("docker daemon is not available: %v: %s", err, strings.TrimSpace(string(output)))
	}

	return binary
}

func requireDockerImage(t *testing.T, binary string, image string) {
	t.Helper()

	command := exec.Command(binary, "image", "inspect", image)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Skipf("docker image %q is not available locally: %v: %s", image, err, strings.TrimSpace(string(output)))
	}
}

func createSubmissionPackage(t *testing.T, workspace, outputPath string) {
	t.Helper()

	file, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("create submission package: %v", err)
	}
	defer file.Close()

	exerciseRef := &exercise.Exercise{Path: workspace}
	if err := exerciseRef.CreateSubmissionPackage(file, "acme", "student-42"); err != nil {
		t.Fatalf("create submission package: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("stat submission package: %v", err)
	}
}

type localArtifactProvider struct {
	path string
}

func (p *localArtifactProvider) OpenPrivateArtifact(_ context.Context, _, _, _ string) (io.ReadCloser, error) {
	return os.Open(p.path)
}

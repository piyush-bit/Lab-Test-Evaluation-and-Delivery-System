package evaluatorcore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var ErrExerciseNotFound = errors.New("exercise not found")

type ArtifactProvider interface {
	OpenPrivateArtifact(ctx context.Context, orgID, labID, version string) (io.ReadCloser, error)
}

type Evaluator struct {
	provider ArtifactProvider
	runtime  Runtime
}

type EvaluationRequest struct {
	SubmissionArchivePath string
	DockerBinary          string
}

type EvaluationResult struct {
	OrgID        string                 `json:"org_id"`
	StudentID    string                 `json:"student_id"`
	LabID        string                 `json:"lab_id"`
	Version      string                 `json:"version"`
	Status       string                 `json:"status"`
	EarnedPoints int                    `json:"earned_points"`
	MaxPoints    int                    `json:"max_points"`
	Results      []EvaluationTestResult `json:"results"`
}

type EvaluationTestResult struct {
	Command        string `json:"command"`
	PointsPossible int    `json:"points_possible"`
	PointsEarned   int    `json:"points_earned"`
	Status         string `json:"status"`
	Output         string `json:"output,omitempty"`
}

func NewEvaluator(provider ArtifactProvider, runtime Runtime) (*Evaluator, error) {
	if provider == nil {
		return nil, fmt.Errorf("artifact provider is required")
	}
	if runtime == nil {
		runtime = NewDockerRuntime()
	}
	return &Evaluator{provider: provider, runtime: runtime}, nil
}

func (e *Evaluator) EvaluateSubmission(ctx context.Context, request EvaluationRequest) (EvaluationResult, error) {
	submissionArchivePath := strings.TrimSpace(request.SubmissionArchivePath)
	if submissionArchivePath == "" {
		return EvaluationResult{Status: "error"}, fmt.Errorf("submission archive path is required")
	}

	submission, err := readSubmissionArchive(submissionArchivePath)
	if err != nil {
		return EvaluationResult{Status: "error"}, err
	}

	result := EvaluationResult{
		OrgID:     submission.Manifest.OrgID,
		StudentID: submission.Manifest.StudentID,
		LabID:     submission.Manifest.LabID,
		Version:   submission.Manifest.Version,
		Status:    "error",
	}

	artifactReader, err := e.provider.OpenPrivateArtifact(
		ctx,
		submission.Manifest.OrgID,
		submission.Manifest.LabID,
		submission.Manifest.Version,
	)
	if err != nil {
		return result, err
	}
	defer artifactReader.Close()

	workspace, err := os.MkdirTemp("", "evaluator-workspace-*")
	if err != nil {
		return result, fmt.Errorf("create workspace: %w", err)
	}
	defer os.RemoveAll(workspace)

	if err := extractExercisePackage(artifactReader, workspace); err != nil {
		return result, fmt.Errorf("extract private exercise package: %w", err)
	}

	manifestPath := filepath.Join(workspace, "manifest.json")
	manifest, err := readManifest(manifestPath)
	if err != nil {
		return result, err
	}
	if err := validateManifest(manifest); err != nil {
		return result, err
	}
	if err := populateReferenceSlot(workspace, manifest, submission); err != nil {
		return result, err
	}

	maxPoints := 0
	for _, entry := range manifest.Grading {
		maxPoints += entry.Points
	}
	result.MaxPoints = maxPoints
	result.Results = make([]EvaluationTestResult, 0, len(manifest.Grading))

	imageConfig := ImageConfig{
		Image:  manifest.RunnerImage,
		Binary: request.DockerBinary,
		Stdout: io.Discard,
		Stderr: io.Discard,
	}

	containerID, err := e.runtime.CreateContainer(ctx, ContainerConfig{
		ImageConfig: imageConfig,
		HostPath:    workspace,
		Workdir:     "/exercise",
		MemoryMB:    manifest.Limits.MemoryMB,
		PidsLimit:   manifest.Limits.PidsLimit,
	})
	if err != nil {
		return result, fmt.Errorf("create grading container: %w", err)
	}
	defer e.runtime.RemoveContainer(context.Background(), imageConfig, containerID)

	if err := e.runtime.StartContainer(ctx, imageConfig, containerID); err != nil {
		return result, fmt.Errorf("start grading container: %w", err)
	}

	timeout := time.Duration(manifest.Limits.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	for _, entry := range manifest.Grading {
		testResult := EvaluationTestResult{
			Command:        entry.Command,
			PointsPossible: entry.Points,
			Status:         "fail",
		}

		var outputBuf bytes.Buffer
		execImageConfig := imageConfig
		execImageConfig.Stdout = &outputBuf
		execImageConfig.Stderr = &outputBuf

		err := e.runtime.ExecCommand(ExecConfig{
			ImageConfig: execImageConfig,
			ContainerID: containerID,
			Command:     entry.Command,
			Timeout:     timeout,
		})

		testResult.Output = strings.TrimSpace(outputBuf.String())

		switch {
		case err == nil:
			testResult.Status = "pass"
			testResult.PointsEarned = entry.Points
			result.EarnedPoints += entry.Points
		case errors.Is(err, context.DeadlineExceeded):
			result.Results = append(result.Results, testResult)
			return result, fmt.Errorf("grading command timed out: %s", entry.Command)
		default:
			var exitErr *ExitError
			if !errors.As(err, &exitErr) {
				result.Results = append(result.Results, testResult)
				return result, fmt.Errorf("run grading command %q: %w", entry.Command, err)
			}
		}

		result.Results = append(result.Results, testResult)
	}

	result.Status = "completed"
	return result, nil
}

func populateReferenceSlot(workspace string, manifest exerciseManifest, submission submissionArchive) error {
	expected := normalizePaths(manifest.Submission.IncludePaths)
	actual := normalizePaths(submission.Manifest.IncludedPaths)
	if !slices.Equal(expected, actual) {
		return fmt.Errorf("submission manifest included_paths do not match exercise submission.include_paths")
	}

	actualFiles := make([]string, 0, len(submission.Files))
	for relPath := range submission.Files {
		actualFiles = append(actualFiles, relPath)
	}
	actualFiles = normalizePaths(actualFiles)
	if !slices.Equal(expected, actualFiles) {
		return fmt.Errorf("submission archive files do not match exercise submission.include_paths")
	}

	for _, relPath := range expected {
		destination, err := safeJoin(workspace, filepath.Join("reference", relPath))
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return fmt.Errorf("create reference directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(destination, submission.Files[relPath], 0644); err != nil {
			return fmt.Errorf("write reference file %s: %w", relPath, err)
		}
	}

	return nil
}

func normalizePaths(paths []string) []string {
	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			continue
		}
		normalized = append(normalized, path)
	}
	slices.Sort(normalized)
	return normalized
}

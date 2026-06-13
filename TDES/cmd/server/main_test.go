package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	evaluatorcore "TDES/internals/evaluator-core"
	"TDES/internals/registry"
)

type mockRuntime struct {
	createContainerFunc func(ctx context.Context, config evaluatorcore.ContainerConfig) (string, error)
	startContainerFunc  func(ctx context.Context, imageConfig evaluatorcore.ImageConfig, containerID string) error
	execCommandFunc     func(config evaluatorcore.ExecConfig) error
	removeContainerFunc func(ctx context.Context, imageConfig evaluatorcore.ImageConfig, containerID string) error
}

func (m *mockRuntime) CreateContainer(ctx context.Context, config evaluatorcore.ContainerConfig) (string, error) {
	if m.createContainerFunc != nil {
		return m.createContainerFunc(ctx, config)
	}
	return "mock-container-id", nil
}

func (m *mockRuntime) StartContainer(ctx context.Context, imageConfig evaluatorcore.ImageConfig, containerID string) error {
	if m.startContainerFunc != nil {
		return m.startContainerFunc(ctx, imageConfig, containerID)
	}
	return nil
}

func (m *mockRuntime) ExecCommand(config evaluatorcore.ExecConfig) error {
	if m.execCommandFunc != nil {
		return m.execCommandFunc(config)
	}
	return nil
}

func (m *mockRuntime) RemoveContainer(ctx context.Context, imageConfig evaluatorcore.ImageConfig, containerID string) error {
	if m.removeContainerFunc != nil {
		return m.removeContainerFunc(ctx, imageConfig, containerID)
	}
	return nil
}

func TestHandleSubmissions(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := registry.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("new sqlite repository: %v", err)
	}
	defer repo.Close()

	store, err := registry.NewDiskArtifactStore(filepath.Join(tempDir, "objects"))
	if err != nil {
		t.Fatalf("new artifact store: %v", err)
	}

	service, err := registry.NewService(repo, store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	// 1. Create a mock manifest for the exercise
	manifestJSON := `{
		"lab_id": "lab1",
		"title": "Test Lab",
		"version": "1.0.0",
		"language": "go",
		"runner_image": "alpine",
		"local_entrypoint": "make test-public",
		"grading": [
			{
				"command": "make test-submission-1",
				"points": 5
			},
			{
				"command": "make test-submission-2",
				"points": 5
			}
		],
		"submission": {
			"include_paths": ["solution.txt"]
		}
	}`

	// 2. Create the mock public package (manifest.json + placeholder files)
	publicTarGz := createMockTarGz(t, map[string]string{
		"manifest.json": manifestJSON,
		"solution.txt":  "// TODO",
	})
	publicPackage := createMockPackage(t, "public.tar.gz", manifestJSON, publicTarGz)

	// 3. Create the mock private package (manifest.json + private tests)
	privateTarGz := createMockTarGz(t, map[string]string{
		"manifest.json": manifestJSON,
		"solution.txt":  "// REFERENCE SOLUTION",
	})
	privatePackage := createMockPackage(t, "private.tar.gz", manifestJSON, privateTarGz)

	// 4. Publish the mock exercise
	_, _, err = service.Publish(context.Background(), registry.PublishRequest{
		OrgID:               "acme",
		ExerciseID:          "lab1",
		Version:             "1.0.0",
		Status:              "published",
		PublicArtifactPath:  publicPackage,
		PrivateArtifactPath: privatePackage,
	})
	if err != nil {
		t.Fatalf("publish test exercise: %v", err)
	}

	// 5. Create a mock student submission archive
	submissionManifestJSON := `{
		"schema_version": "v1",
		"org_id": "acme",
		"student_id": "student-1",
		"lab_id": "lab1",
		"version": "1.0.0",
		"created_at": "2026-06-13T12:00:00Z",
		"included_paths": ["solution.txt"]
	}`
	submissionPackagePath := createMockTar(t, "submission-manifest.json", []byte(submissionManifestJSON))
	// Add solution.txt to the submission package
	appendFileToTar(t, submissionPackagePath, "solution.txt", "student solution")

	// Set up mock runtime
	executedCommands := make([]string, 0)
	runtime := &mockRuntime{
		execCommandFunc: func(config evaluatorcore.ExecConfig) error {
			executedCommands = append(executedCommands, config.Command)
			if config.Command == "make test-submission-2" {
				return &evaluatorcore.ExitError{Code: 1} // simulate one fail
			}
			return nil // simulate pass
		},
	}

	// 6. Test calling POST /v1/submissions
	handler := handleSubmissions(service, runtime)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Prepare multipart request body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("submission_package", "submission.tar")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	fileBytes, err := os.ReadFile(submissionPackagePath)
	if err != nil {
		t.Fatalf("read submission package: %v", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		t.Fatalf("write form file body: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 OK, got %d. Body: %s", resp.StatusCode, string(respBody))
	}

	var result evaluatorcore.EvaluationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode evaluation result JSON: %v", err)
	}

	if result.OrgID != "acme" || result.StudentID != "student-1" || result.LabID != "lab1" {
		t.Errorf("unexpected submission metadata in result: %+v", result)
	}
	if result.EarnedPoints != 5 || result.MaxPoints != 10 {
		t.Errorf("unexpected score: earned=%d max=%d", result.EarnedPoints, result.MaxPoints)
	}
	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}
	if result.Results[0].Status != "pass" || result.Results[1].Status != "fail" {
		t.Errorf("unexpected results detail: %+v", result.Results)
	}

	if len(executedCommands) != 2 || executedCommands[0] != "make test-submission-1" || executedCommands[1] != "make test-submission-2" {
		t.Errorf("unexpected grading commands executed: %v", executedCommands)
	}
}

func TestHandleSubmissionsTokenCheck(t *testing.T) {
	t.Setenv("EUC2_REMOTE_BEARER_TOKEN", "valid-token")

	handler := handleSubmissions(nil, nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}

// helper functions for creating mock files/archives in tests

func createMockTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write tar body: %v", err)
		}
	}

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func createMockPackage(t *testing.T, filename string, manifestJSON string, gzipData []byte) string {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "mock-package-*.tar")
	if err != nil {
		t.Fatalf("create temp tar file: %v", err)
	}
	defer tempFile.Close()

	tw := tar.NewWriter(tempFile)
	manifestHeader := &tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestJSON)),
	}
	if err := tw.WriteHeader(manifestHeader); err != nil {
		t.Fatalf("write manifest.json header to tar: %v", err)
	}
	if _, err := tw.Write([]byte(manifestJSON)); err != nil {
		t.Fatalf("write manifest.json data to tar: %v", err)
	}

	gzipHeader := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(gzipData)),
	}
	if err := tw.WriteHeader(gzipHeader); err != nil {
		t.Fatalf("write inner archive header to tar: %v", err)
	}
	if _, err := tw.Write(gzipData); err != nil {
		t.Fatalf("write inner archive data to tar: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	return tempFile.Name()
}

func createMockTar(t *testing.T, filename string, data []byte) string {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "mock-archive-*.tar")
	if err != nil {
		t.Fatalf("create temp tar file: %v", err)
	}
	defer tempFile.Close()

	tw := tar.NewWriter(tempFile)
	header := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write header to tar: %v", err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatalf("write data to tar: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	return tempFile.Name()
}

func appendFileToTar(t *testing.T, tarPath string, name string, content string) {
	t.Helper()
	// Read existing tar
	data, err := os.ReadFile(tarPath)
	if err != nil {
		t.Fatalf("read tar file: %v", err)
	}

	// Create new tar with existing entries + new entry
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Copy existing entries (omitting EOF)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read entry: %v", err)
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := io.Copy(tw, tr); err != nil {
			t.Fatalf("copy body: %v", err)
		}
	}

	// Add new entry
	header := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write new header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("write new body: %v", err)
	}

	tw.Close()

	// Overwrite original file
	if err := os.WriteFile(tarPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("overwrite tar file: %v", err)
	}
}

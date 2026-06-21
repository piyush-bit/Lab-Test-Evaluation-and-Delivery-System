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

	// Onboard student to roster
	err = service.SaveStudentCredential(context.Background(), registry.StudentCredential{
		OrgID:     "acme",
		StudentID: "student-1",
	})
	if err != nil {
		t.Fatalf("onboard student: %v", err)
	}

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

	// Add pin field
	if err := writer.WriteField("pin", "1234"); err != nil {
		t.Fatalf("write pin form field: %v", err)
	}

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

	// 7. Verify database entry was persisted
	count, err := repo.GetSubmissionsCountForTesting(context.Background())
	if err != nil {
		t.Fatalf("get submissions count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 submission record in database, got %d", count)
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

func TestGetSubmissions(t *testing.T) {
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

	// Set test token
	const testToken = "super-secret-token"
	t.Setenv("EUC2_REMOTE_BEARER_TOKEN", testToken)

	// Seed database
	ctx := context.Background()
	_ = service.SaveEvaluation(ctx, registry.SubmissionEvaluation{
		ID:           "eval-1",
		OrgID:        "org-acme",
		StudentID:    "student-1",
		LabID:        "lab-1",
		Version:      "1.0",
		Status:       "completed",
		EarnedPoints: 9,
		MaxPoints:    10,
		ResultsJSON:  "[]",
	})
	_ = service.SaveEvaluation(ctx, registry.SubmissionEvaluation{
		ID:           "eval-2",
		OrgID:        "org-acme",
		StudentID:    "student-2",
		LabID:        "lab-2",
		Version:      "1.0",
		Status:       "completed",
		EarnedPoints: 5,
		MaxPoints:    10,
		ResultsJSON:  "[]",
	})

	handler := handleGetSubmissions(service)

	// 1. Test unauthorized request
	req, _ := http.NewRequest("GET", "/v1/submissions", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}

	// 2. Test authorized request (JSON by default)
	req, _ = http.NewRequest("GET", "/v1/submissions", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var jsonResults []registry.SubmissionEvaluation
	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResults); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if len(jsonResults) != 2 {
		t.Errorf("expected 2 submissions, got %d", len(jsonResults))
	}

	// 3. Test authorized request (CSV)
	req, _ = http.NewRequest("GET", "/v1/submissions?format=csv", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", contentType)
	}
	bodyStr := rr.Body.String()
	if !bytes.Contains([]byte(bodyStr), []byte("student-1")) {
		t.Errorf("expected body to contain student-1, got %q", bodyStr)
	}
	if !bytes.Contains([]byte(bodyStr), []byte("student-2")) {
		t.Errorf("expected body to contain student-2, got %q", bodyStr)
	}
	if !bytes.Contains([]byte(bodyStr), []byte("id,org_id,student_id,lab_id,version,status,earned_points,max_points,created_at")) {
		t.Errorf("expected body to contain CSV header, got %q", bodyStr)
	}
}

func TestAdminOnboard(t *testing.T) {
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

	const testToken = "admin-secret-token"
	t.Setenv("EUC2_INSTRUCTOR_TOKENS", testToken)

	handler := handleAdminOnboard(service)

	// 1. Test unauthorized request
	req, _ := http.NewRequest("POST", "/v1/admin/onboard", bytes.NewReader([]byte("student_id,org_id\nstudent-1,acme")))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}

	// 2. Test authorized request (valid CSV body)
	req, _ = http.NewRequest("POST", "/v1/admin/onboard", bytes.NewReader([]byte("student_id,org_id\nstudent-1,acme\nstudent-2,acme")))
	req.Header.Set("Authorization", "Bearer "+testToken)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verify database state
	cred1, err := service.GetStudentCredential(context.Background(), "acme", "student-1")
	if err != nil {
		t.Fatalf("student-1 not found in roster: %v", err)
	}
	if cred1.PinHash != "" {
		t.Errorf("expected empty pin_hash, got %q", cred1.PinHash)
	}

	cred2, err := service.GetStudentCredential(context.Background(), "acme", "student-2")
	if err != nil {
		t.Fatalf("student-2 not found in roster: %v", err)
	}
	if cred2.PinHash != "" {
		t.Errorf("expected empty pin_hash, got %q", cred2.PinHash)
	}
}

func TestPINVerificationFlow(t *testing.T) {
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

	// Create mock manifest and package for remote submissions
	manifestJSON := `{
		"lab_id": "lab1",
		"title": "Test Lab",
		"version": "1.0.0",
		"language": "go",
		"runner_image": "alpine",
		"local_entrypoint": "make test-public",
		"grading": [{"command": "make test-submission-1", "points": 10}],
		"submission": {"include_paths": ["solution.txt"]}
	}`
	publicTarGz := createMockTarGz(t, map[string]string{
		"manifest.json": manifestJSON,
		"solution.txt":  "// TODO",
	})
	publicPackage := createMockPackage(t, "public.tar.gz", manifestJSON, publicTarGz)
	privateTarGz := createMockTarGz(t, map[string]string{
		"manifest.json": manifestJSON,
		"solution.txt":  "// REFERENCE",
	})
	privatePackage := createMockPackage(t, "private.tar.gz", manifestJSON, privateTarGz)

	_, _, _ = service.Publish(context.Background(), registry.PublishRequest{
		OrgID:               "acme",
		ExerciseID:          "lab1",
		Version:             "1.0.0",
		PublicArtifactPath:  publicPackage,
		PrivateArtifactPath: privatePackage,
	})

	// Pre-register roster ID
	_ = service.SaveStudentCredential(context.Background(), registry.StudentCredential{
		OrgID:     "acme",
		StudentID: "rostered-student",
	})

	// Prepare mock runtime
	runtime := &mockRuntime{
		execCommandFunc: func(config evaluatorcore.ExecConfig) error {
			return nil
		},
	}
	handler := handleSubmissions(service, runtime)

	// Create a valid submission package with student-manifest
	submissionManifestJSON := `{
		"schema_version": "v1",
		"org_id": "acme",
		"student_id": "rostered-student",
		"lab_id": "lab1",
		"version": "1.0.0",
		"created_at": "2026-06-13T12:00:00Z",
		"included_paths": ["solution.txt"]
	}`
	submissionPackagePath := createMockTar(t, "submission-manifest.json", []byte(submissionManifestJSON))
	appendFileToTar(t, submissionPackagePath, "solution.txt", "student solution")

	// Helper to send a remote submission request
	sendSubmission := func(studentID, pin, newPin string) (int, string) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		if pin != "" {
			_ = writer.WriteField("pin", pin)
		}
		if newPin != "" {
			_ = writer.WriteField("new_pin", newPin)
		}

		part, _ := writer.CreateFormFile("submission_package", "submission.tar")
		fileBytes, _ := os.ReadFile(submissionPackagePath)
		_, _ = part.Write(fileBytes)
		_ = writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/v1/submissions", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr.Code, rr.Body.String()
	}

	// 1. Submit as a non-rostered student
	nonRosteredManifest := `{
		"schema_version": "v1",
		"org_id": "acme",
		"student_id": "fake-student",
		"lab_id": "lab1",
		"version": "1.0.0",
		"created_at": "2026-06-13T12:00:00Z",
		"included_paths": ["solution.txt"]
	}`
	submissionPackagePathFake := createMockTar(t, "submission-manifest.json", []byte(nonRosteredManifest))
	appendFileToTar(t, submissionPackagePathFake, "solution.txt", "student solution")

	bodyFake := &bytes.Buffer{}
	writerFake := multipart.NewWriter(bodyFake)
	_ = writerFake.WriteField("pin", "1234")
	partFake, _ := writerFake.CreateFormFile("submission_package", "submission.tar")
	fileBytesFake, _ := os.ReadFile(submissionPackagePathFake)
	_, _ = partFake.Write(fileBytesFake)
	_ = writerFake.Close()

	reqFake, _ := http.NewRequest(http.MethodPost, "/v1/submissions", bodyFake)
	reqFake.Header.Set("Content-Type", writerFake.FormDataContentType())
	rrFake := httptest.NewRecorder()
	handler.ServeHTTP(rrFake, reqFake)
	if rrFake.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-rostered student, got %d. Body: %s", rrFake.Code, rrFake.Body.String())
	}

	// 2. Submit as rostered student with too short PIN (TOFU activation should fail)
	code, bodyStr := sendSubmission("rostered-student", "12", "")
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for short PIN, got %d. Body: %s", code, bodyStr)
	}

	// 3. Submit with valid PIN (TOFU activation succeeds)
	code, bodyStr = sendSubmission("rostered-student", "123456", "")
	if code != http.StatusOK {
		t.Errorf("expected 200 OK for valid PIN TOFU activation, got %d. Body: %s", code, bodyStr)
	}

	// 4. Submit again with incorrect PIN (should fail)
	code, bodyStr = sendSubmission("rostered-student", "wrong-pin", "")
	if code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for incorrect PIN, got %d. Body: %s", code, bodyStr)
	}

	// 5. Submit again with correct PIN and a new PIN (should succeed and update PIN)
	code, bodyStr = sendSubmission("rostered-student", "123456", "654321")
	if code != http.StatusOK {
		t.Errorf("expected 200 OK for PIN update, got %d. Body: %s", code, bodyStr)
	}

	// 6. Submit again with old PIN (should fail now)
	code, bodyStr = sendSubmission("rostered-student", "123456", "")
	if code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for old PIN, got %d. Body: %s", code, bodyStr)
	}

	// 7. Submit again with new PIN (should succeed)
	code, bodyStr = sendSubmission("rostered-student", "654321", "")
	if code != http.StatusOK {
		t.Errorf("expected 200 OK for new PIN, got %d. Body: %s", code, bodyStr)
	}
}



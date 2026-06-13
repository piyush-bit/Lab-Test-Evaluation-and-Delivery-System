package remote

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercisestore "TDES/internals/exercise_store"
)

func TestFetchFromRemoteSavesPackageToCache(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("EUC2_CACHE_DIR", cacheDir)

	artifactBytes := buildExercisePackageBytes(t, map[string]any{
		"lab_id":   "lab-1",
		"version":  "v1",
		"title":    "Remote Lab",
		"language": "go",
	})

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/v1/exercises/org-1/lab-1/versions/v1":
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(ExerciseVersion{
				OrgID:             "org-1",
				ExerciseID:        "lab-1",
				Version:           "v1",
				PublicArtifactSHA: "artifact-sha",
			})
		case "/v1/artifacts/artifact-sha":
			writer.Header().Set("Content-Type", "application/octet-stream")
			_, _ = writer.Write(artifactBytes)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	remoteRef := NewRemote(server.URL)
	if err := remoteRef.FetchFromRemote("lab-1", "v1", "org-1"); err != nil {
		t.Fatalf("fetch from remote: %v", err)
	}

	packagePath, err := exercisestore.ResolveExercisePath(exercisestore.GetPublicCacheDir(), "lab-1", "v1")
	if err != nil {
		t.Fatalf("resolve cached package: %v", err)
	}
	if _, err := os.Stat(packagePath); err != nil {
		t.Fatalf("stat cached package: %v", err)
	}
}

func TestSubmitRemoteUploadsSubmissionPackageWithBearerToken(t *testing.T) {
	exerciseDir := t.TempDir()
	writeFile(t, filepath.Join(exerciseDir, "manifest.json"), `{
  "lab_id": "lab-1",
  "title": "Remote Lab",
  "version": "v1",
  "language": "go",
  "runner_image": "golang:1.24",
  "local_entrypoint": "make test-public",
  "grading": [
    { "command": "go test ./...", "points": 100 }
  ],
  "submission": {
    "include_paths": ["solution.txt"],
    "private_globs": [],
    "exclude_globs": []
  },
  "limits": {
    "memory_mb": 128,
    "timeout_seconds": 30,
    "pids_limit": 64
  }
}`)
	writeFile(t, filepath.Join(exerciseDir, "solution.txt"), "hello remote\n")

	var (
		seenAuthHeader string
		seenFileName   string
		seenFileBytes  []byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/submissions" {
			http.NotFound(writer, request)
			return
		}
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", request.Method)
		}

		seenAuthHeader = request.Header.Get("Authorization")
		reader, err := request.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			if part.FormName() != "submission_package" {
				continue
			}
			seenFileName = part.FileName()
			seenFileBytes, err = io.ReadAll(part)
			if err != nil {
				t.Fatalf("read part: %v", err)
			}
		}

		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte("accepted"))
	}))
	defer server.Close()

	remoteRef := NewRemote(server.URL)
	message, err := remoteRef.SubmitRemote(SubmitRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-1",
		StudentID:    "student-1",
		BearerToken:  "secret-token",
	})
	if err != nil {
		t.Fatalf("submit remote: %v", err)
	}

	if message != "accepted" {
		t.Fatalf("unexpected response %q", message)
	}
	if seenAuthHeader != "Bearer secret-token" {
		t.Fatalf("unexpected auth header %q", seenAuthHeader)
	}
	if strings.TrimSpace(seenFileName) == "" {
		t.Fatalf("submission package filename was not sent")
	}

	manifest, err := readSubmissionManifest(t, seenFileBytes)
	if err != nil {
		t.Fatalf("read submission manifest: %v", err)
	}
	if manifest.OrgID != "org-1" {
		t.Fatalf("unexpected org id %q", manifest.OrgID)
	}
	if manifest.StudentID != "student-1" {
		t.Fatalf("unexpected student id %q", manifest.StudentID)
	}
}

func TestSubmitRemoteUsesBearerTokenFromEnvironment(t *testing.T) {
	exerciseDir := t.TempDir()
	t.Setenv(BearerTokenEnvVar, "env-secret-token")

	writeFile(t, filepath.Join(exerciseDir, "manifest.json"), `{
  "lab_id": "lab-1",
  "title": "Remote Lab",
  "version": "v1",
  "language": "go",
  "runner_image": "golang:1.24",
  "local_entrypoint": "make test-public",
  "grading": [
    { "command": "go test ./...", "points": 100 }
  ],
  "submission": {
    "include_paths": ["solution.txt"],
    "private_globs": [],
    "exclude_globs": []
  },
  "limits": {
    "memory_mb": 128,
    "timeout_seconds": 30,
    "pids_limit": 64
  }
}`)
	writeFile(t, filepath.Join(exerciseDir, "solution.txt"), "hello remote\n")

	var seenAuthHeader string

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		seenAuthHeader = request.Header.Get("Authorization")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte("accepted"))
	}))
	defer server.Close()

	remoteRef := NewRemote(server.URL)
	_, err := remoteRef.SubmitRemote(SubmitRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-1",
		StudentID:    "student-1",
	})
	if err != nil {
		t.Fatalf("submit remote with env token: %v", err)
	}

	if seenAuthHeader != "Bearer env-secret-token" {
		t.Fatalf("unexpected auth header %q", seenAuthHeader)
	}
}

func buildExercisePackageBytes(t *testing.T, manifest map[string]any) []byte {
	t.Helper()

	buffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buffer)

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestBytes)),
	}); err != nil {
		t.Fatalf("write manifest header: %v", err)
	}
	if _, err := tarWriter.Write(manifestBytes); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}

	return buffer.Bytes()
}

func readSubmissionManifest(t *testing.T, archiveBytes []byte) (*submissionArchiveManifest, error) {
	t.Helper()

	reader := tar.NewReader(bytes.NewReader(archiveBytes))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("submission manifest not found")
		}
		if err != nil {
			return nil, err
		}
		if header.Name != "submission-manifest.json" {
			continue
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		var manifest submissionArchiveManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, err
		}
		return &manifest, nil
	}
}

type submissionArchiveManifest struct {
	OrgID      string `json:"org_id"`
	StudentID  string `json:"student_id"`
	LabID      string `json:"lab_id"`
	Version    string `json:"version"`
	CreatedAt  string `json:"created_at"`
	SchemaName string `json:"schema_version"`
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestSubmissionMultipartUsesExpectedFieldName(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	part, err := writer.CreateFormFile("submission_package", "submission.tar")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("payload")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/submissions", buffer)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	reader, err := request.MultipartReader()
	if err != nil {
		t.Fatalf("multipart reader: %v", err)
	}
	partReader, err := reader.NextPart()
	if err != nil {
		t.Fatalf("next part: %v", err)
	}
	if partReader.FormName() != "submission_package" {
		t.Fatalf("unexpected form field %q", partReader.FormName())
	}
}

func TestPublishRemoteUploadsCorrectly(t *testing.T) {
	var (
		seenOrgID          string
		seenExID           string
		seenVer            string
		seenStat           string
		seenPublicContent  string
		seenPrivateContent string
		seenAuthHeader     string
	)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/exercises/publish" {
			http.NotFound(writer, request)
			return
		}
		if request.Method != http.MethodPost {
			http.Error(writer, "expected POST", http.StatusMethodNotAllowed)
			return
		}

		seenAuthHeader = request.Header.Get("Authorization")

		if err := request.ParseMultipartForm(10 << 20); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		seenOrgID = request.FormValue("org_id")
		seenExID = request.FormValue("exercise_id")
		seenVer = request.FormValue("version")
		seenStat = request.FormValue("status")

		publicPart, _, err := request.FormFile("public_artifact")
		if err == nil {
			defer publicPart.Close()
			buf := new(bytes.Buffer)
			_, _ = io.Copy(buf, publicPart)
			seenPublicContent = buf.String()
		}

		privatePart, _, err := request.FormFile("private_artifact")
		if err == nil {
			defer privatePart.Close()
			buf := new(bytes.Buffer)
			_, _ = io.Copy(buf, privatePart)
			seenPrivateContent = buf.String()
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("mock-publish-ok"))
	}))
	defer server.Close()

	remoteRef := NewRemote(server.URL)
	tempDir := t.TempDir()

	// Mock the packager function to return dummy file paths containing simple contents
	remoteRef.PackageFunc = func(path string) (string, string, error) {
		publicFile := filepath.Join(tempDir, "mock-pub.tar")
		privateFile := filepath.Join(tempDir, "mock-priv.tar")

		if err := os.WriteFile(publicFile, []byte("mock public archive"), 0644); err != nil {
			return "", "", err
		}
		if err := os.WriteFile(privateFile, []byte("mock private archive"), 0644); err != nil {
			return "", "", err
		}
		return publicFile, privateFile, nil
	}

	response, err := remoteRef.PublishRemote(PublishRequest{
		ExercisePath: tempDir,
		OrgID:        "test-org",
		ExerciseID:   "test-ex",
		Version:      "2.0.0",
		Status:       "draft",
		BearerToken:  "test-token",
	})
	if err != nil {
		t.Fatalf("PublishRemote failed: %v", err)
	}

	if response != "mock-publish-ok" {
		t.Errorf("unexpected response: %q", response)
	}
	if seenOrgID != "test-org" || seenExID != "test-ex" || seenVer != "2.0.0" || seenStat != "draft" {
		t.Errorf("incorrect form fields received: org=%q, ex=%q, ver=%q, stat=%q", seenOrgID, seenExID, seenVer, seenStat)
	}
	if seenPublicContent != "mock public archive" || seenPrivateContent != "mock private archive" {
		t.Errorf("incorrect file content received: pub=%q, priv=%q", seenPublicContent, seenPrivateContent)
	}
	if seenAuthHeader != "Bearer test-token" {
		t.Errorf("unexpected authorization header: %q", seenAuthHeader)
	}
}

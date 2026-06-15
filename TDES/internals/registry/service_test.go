package registry

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func createMockPackage(t *testing.T, dir string, filename string, isGzip bool, manifestJSON string) string {
	tmpPath := filepath.Join(dir, filename)
	file, err := os.Create(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var writer io.Writer = file
	var gw *gzip.Writer
	if isGzip {
		gw = gzip.NewWriter(file)
		writer = gw
	}

	tw := tar.NewWriter(writer)
	
	header := &tar.Header{
		Name: "manifest.json",
		Mode: 0644,
		Size: int64(len(manifestJSON)),
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(manifestJSON)); err != nil {
		t.Fatal(err)
	}
	
	tw.Close()
	if gw != nil {
		gw.Close()
	}

	return tmpPath
}

func TestRegistryService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tdes-registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	storePath := filepath.Join(tempDir, "objects")
	store, err := NewDiskArtifactStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repo, store)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// 1. Create mock package manifests
	manifestJSON := `{
		"lab_id": "lab01",
		"version": "1.0.0",
		"title": "Lab 1",
		"language": "go",
		"runner_image": "golang:1.24",
		"local_entrypoint": "make test-public",
		"grading": [{"command": "go test", "points": 100}],
		"submission": {"include_paths": ["main.go"]},
		"limits": {"memory_mb": 512, "timeout_seconds": 10, "pids_limit": 100}
	}`

	publicPkg := createMockPackage(t, tempDir, "public.tar.gz", true, manifestJSON)
	privatePkg := createMockPackage(t, tempDir, "private.tar", false, manifestJSON)

	// 2. Publish
	req := PublishRequest{
		OrgID:               "org1",
		ExerciseID:          "lab01",
		Version:             "1.0.0",
		Status:              "draft",
		PublicArtifactPath:  publicPkg,
		PrivateArtifactPath: privatePkg,
	}

	storedVersion, created, err := service.Publish(ctx, req)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if !created {
		t.Errorf("Expected created to be true")
	}

	if storedVersion.OrgID != "org1" || storedVersion.ExerciseID != "lab01" || storedVersion.Version != "1.0.0" {
		t.Errorf("Unexpected stored version: %+v", storedVersion)
	}

	// 3. Re-publishing identical should succeed but return created=false
	_, created, err = service.Publish(ctx, req)
	if err != nil {
		t.Fatalf("Publish duplicate failed: %v", err)
	}
	if created {
		t.Errorf("Expected created to be false for duplicate publish")
	}

	// 4. Re-publishing with different metadata should conflict
	reqConflicting := req
	reqConflicting.Status = "published"
	_, _, err = service.Publish(ctx, reqConflicting)
	if err == nil || err != ErrExerciseVersionConflict {
		t.Errorf("Expected ErrExerciseVersionConflict, got: %v", err)
	}

	// 5. Get Exercise Version
	ev, err := service.GetExerciseVersion(ctx, "org1", "lab01", "1.0.0")
	if err != nil {
		t.Fatalf("GetExerciseVersion failed: %v", err)
	}
	if ev.Title != "Lab 1" || ev.Status != "draft" {
		t.Errorf("Unexpected fetched version: %+v", ev)
	}

	// 6. Open Artifact
	handle, err := service.OpenArtifact(ctx, ev.PublicArtifactSHA)
	if err != nil {
		t.Fatalf("OpenArtifact failed: %v", err)
	}
	defer handle.File.Close()

	content, err := io.ReadAll(handle.File)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) == 0 {
		t.Error("Artifact stream was empty")
	}

	// 7. Update status
	err = service.UpdateExerciseStatus(ctx, "org1", "lab01", "1.0.0", "published")
	if err != nil {
		t.Fatalf("UpdateExerciseStatus failed: %v", err)
	}

	ev, _ = service.GetExerciseVersion(ctx, "org1", "lab01", "1.0.0")
	if ev.Status != "published" {
		t.Errorf("Expected status 'published', got %s", ev.Status)
	}

	// 8. List exercises
	list, err := service.ListExercises(ctx, "org1", "published")
	if err != nil {
		t.Fatalf("ListExercises failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 exercise in list, got %d", len(list))
	}

	listEmpty, _ := service.ListExercises(ctx, "org1", "draft")
	if len(listEmpty) != 0 {
		t.Errorf("Expected 0 exercises in list, got %d", len(listEmpty))
	}

	// 9. List versions
	versions, err := service.ListExerciseVersions(ctx, "org1", "lab01")
	if err != nil {
		t.Fatalf("ListExerciseVersions failed: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	// 10. Delete version
	err = service.DeleteExerciseVersion(ctx, "org1", "lab01", "1.0.0")
	if err != nil {
		t.Fatalf("DeleteExerciseVersion failed: %v", err)
	}

	_, err = service.GetExerciseVersion(ctx, "org1", "lab01", "1.0.0")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after deletion, got %v", err)
	}
}

func TestSaveEvaluation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tdes-evaluation-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	storePath := filepath.Join(tempDir, "objects")
	store, err := NewDiskArtifactStore(storePath)
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(repo, store)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	eval := SubmissionEvaluation{
		ID:           "test-eval-id",
		OrgID:        "acme",
		StudentID:    "student-1",
		LabID:        "go101-lab01",
		Version:      "1.1.0",
		Status:       "completed",
		EarnedPoints: 8,
		MaxPoints:    10,
		ResultsJSON:  `[{"command":"make test-submission-1","points_possible":5,"points_earned":5,"status":"pass"}]`,
	}

	err = service.SaveEvaluation(ctx, eval)
	if err != nil {
		t.Fatalf("SaveEvaluation failed: %v", err)
	}

	// Verify database record by directly querying the repo's db
	var count int
	err = repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM submissions WHERE id = ?", "test-eval-id").Scan(&count)
	if err != nil {
		t.Fatalf("query submissions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 submission record, got %d", count)
	}

	var storedEarnedPoints int
	err = repo.db.QueryRowContext(ctx, "SELECT earned_points FROM submissions WHERE id = ?", "test-eval-id").Scan(&storedEarnedPoints)
	if err != nil {
		t.Fatalf("query earned_points: %v", err)
	}
	if storedEarnedPoints != 8 {
		t.Errorf("Expected earned_points to be 8, got %d", storedEarnedPoints)
	}
}

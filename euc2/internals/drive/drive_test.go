package drive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exercisestore "euc2/internals/exercise_store"
)

func TestPrepareDriveCreatesManifestAndExerciseStore(t *testing.T) {
	driveRoot := t.TempDir()

	if err := PrepareDrive(driveRoot); err != nil {
		t.Fatalf("PrepareDrive failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(driveRoot, "manifest.json")); err != nil {
		t.Fatalf("expected manifest.json to be created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(driveRoot, "exercise", "index.json")); err != nil {
		t.Fatalf("expected exercise index to be created: %v", err)
	}
}

func TestPrepareDriveForSubmissionCreatesSubmissionManifest(t *testing.T) {
	driveRoot := t.TempDir()
	recipientPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}

	if err := PrepareDriveForSubmission(driveRoot, base64.StdEncoding.EncodeToString(recipientPrivateKey.PublicKey().Bytes())); err != nil {
		t.Fatalf("PrepareDriveForSubmission failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(driveRoot, "submissions", "manifest.json"))
	if err != nil {
		t.Fatalf("read submissions manifest: %v", err)
	}

	var manifest SubmissionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode submission manifest: %v", err)
	}
	if manifest.SchemaVersion != submissionManifestSchema {
		t.Fatalf("expected schema version %q, got %q", submissionManifestSchema, manifest.SchemaVersion)
	}
	if manifest.EncryptAlg != submissionEncryptAlgorithm {
		t.Fatalf("expected encrypt alg %q, got %q", submissionEncryptAlgorithm, manifest.EncryptAlg)
	}
}

func TestDriveAddExerciseFromFileAndResolve(t *testing.T) {
	driveRoot := t.TempDir()
	if err := PrepareDrive(driveRoot); err != nil {
		t.Fatalf("PrepareDrive failed: %v", err)
	}

	packagePath := filepath.Join(t.TempDir(), "exercise.tar")
	if err := writeExercisePackage(packagePath, "go101-lab01", "1.0.0"); err != nil {
		t.Fatalf("write exercise package: %v", err)
	}

	drive := &Drive{Path: driveRoot}
	if err := drive.AddExerciseFromFile(packagePath); err != nil {
		t.Fatalf("AddExerciseFromFile failed: %v", err)
	}

	resolvedPath, err := drive.GetExerciseFile("go101-lab01", "")
	if err != nil {
		t.Fatalf("GetExerciseFile failed: %v", err)
	}
	if _, err := os.Stat(resolvedPath); err != nil {
		t.Fatalf("expected resolved exercise path to exist: %v", err)
	}
}

func TestDriveAddExerciseFromIDUsesLocalCache(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("EUC2_CACHE_DIR", cacheDir)

	packagePath := filepath.Join(t.TempDir(), "exercise.tar")
	if err := writeExercisePackage(packagePath, "go101-lab02", "2.0.0"); err != nil {
		t.Fatalf("write exercise package: %v", err)
	}
	if err := exercisestore.SavePackage(exercisestore.GetPublicCacheDir(), packagePath); err != nil {
		t.Fatalf("save package to cache failed: %v", err)
	}

	driveRoot := t.TempDir()
	if err := PrepareDrive(driveRoot); err != nil {
		t.Fatalf("PrepareDrive failed: %v", err)
	}

	drive := &Drive{Path: driveRoot}
	if err := drive.AddExerciseFromID("go101-lab02", "2.0.0"); err != nil {
		t.Fatalf("AddExerciseFromID failed: %v", err)
	}

	resolvedPath, err := drive.GetExerciseFile("go101-lab02", "2.0.0")
	if err != nil {
		t.Fatalf("GetExerciseFile failed: %v", err)
	}
	if _, err := os.Stat(resolvedPath); err != nil {
		t.Fatalf("expected resolved exercise path to exist: %v", err)
	}
}

func TestDriveCreateSubmissionEncryptsArchiveAndStoresItInExerciseDirectory(t *testing.T) {
	recipientPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}

	driveRoot := t.TempDir()
	if err := PrepareDriveForSubmission(driveRoot, base64.StdEncoding.EncodeToString(recipientPrivateKey.PublicKey().Bytes())); err != nil {
		t.Fatalf("PrepareDriveForSubmission failed: %v", err)
	}

	exerciseDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(exerciseDir, "manifest.json"), []byte(`{
  "lab_id": "go101-lab01",
  "title": "Slice-Backed Stack in Go",
  "version": "1.0.0",
  "language": "go",
  "submission": {
    "include_paths": ["stack.go"]
  }
}`), 0644); err != nil {
		t.Fatalf("write exercise manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(exerciseDir, "stack.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write submission file: %v", err)
	}

	driveRef := &Drive{Path: driveRoot}
	submissionPath, err := driveRef.CreateSubmission(SubmissionRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-1",
		StudentID:    "student-42",
	})
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}

	expectedDir := filepath.Join(driveRoot, "submissions", "go101-lab01")
	if filepath.Dir(submissionPath) != expectedDir {
		t.Fatalf("expected submission in %q, got %q", expectedDir, submissionPath)
	}
	if filepath.Ext(submissionPath) != ".json" {
		t.Fatalf("expected JSON submission envelope, got %q", submissionPath)
	}

	envelopeData, err := os.ReadFile(submissionPath)
	if err != nil {
		t.Fatalf("read submission envelope: %v", err)
	}

	var envelope SubmissionEnvelope
	if err := json.Unmarshal(envelopeData, &envelope); err != nil {
		t.Fatalf("decode submission envelope: %v", err)
	}
	if envelope.SchemaVersion != submissionEnvelopeSchema {
		t.Fatalf("expected schema version %q, got %q", submissionEnvelopeSchema, envelope.SchemaVersion)
	}
	plaintext, err := DecryptSubmissionArchive(envelope, recipientPrivateKey)
	if err != nil {
		t.Fatalf("decrypt submission archive: %v", err)
	}

	archivePath := filepath.Join(t.TempDir(), "submission.tar")
	if err := os.WriteFile(archivePath, plaintext, 0644); err != nil {
		t.Fatalf("write decrypted archive: %v", err)
	}

	entryOrder, entries, err := readTarEntries(archivePath)
	if err != nil {
		t.Fatalf("read decrypted tar entries: %v", err)
	}
	if len(entryOrder) != 2 {
		t.Fatalf("expected 2 tar entries, got %d", len(entryOrder))
	}
	if entryOrder[0] != "submission-manifest.json" {
		t.Fatalf("expected first tar entry submission-manifest.json, got %q", entryOrder[0])
	}
	if entryOrder[1] != "stack.go" {
		t.Fatalf("expected second tar entry stack.go, got %q", entryOrder[1])
	}

	if string(entries["stack.go"]) != "package main\n" {
		t.Fatalf("unexpected stack.go content: %q", string(entries["stack.go"]))
	}
	if !strings.Contains(string(entries["submission-manifest.json"]), `"org_id": "org-1"`) {
		t.Fatalf("expected org id in submission manifest, got %q", string(entries["submission-manifest.json"]))
	}
	if !strings.Contains(string(entries["submission-manifest.json"]), `"student_id": "student-42"`) {
		t.Fatalf("expected student id in submission manifest, got %q", string(entries["submission-manifest.json"]))
	}
	if strings.Contains(string(envelopeData), `"exercise_id"`) || strings.Contains(string(envelopeData), `"org_id"`) || strings.Contains(string(envelopeData), `"student_id"`) {
		t.Fatalf("expected submission details to stay inside encrypted archive, got %q", string(envelopeData))
	}
}

func TestDriveCreateSubmissionKeepsMultipleSubmissionFiles(t *testing.T) {
	recipientPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}

	driveRoot := t.TempDir()
	if err := PrepareDriveForSubmission(driveRoot, base64.StdEncoding.EncodeToString(recipientPrivateKey.PublicKey().Bytes())); err != nil {
		t.Fatalf("PrepareDriveForSubmission failed: %v", err)
	}

	exerciseDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(exerciseDir, "manifest.json"), []byte(`{
  "lab_id": "go101-lab01",
  "title": "Slice-Backed Stack in Go",
  "version": "1.0.0",
  "language": "go",
  "submission": {
    "include_paths": ["stack.go"]
  }
}`), 0644); err != nil {
		t.Fatalf("write exercise manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(exerciseDir, "stack.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write submission file: %v", err)
	}

	driveRef := &Drive{Path: driveRoot}
	firstPath, err := driveRef.CreateSubmission(SubmissionRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-1",
		StudentID:    "student-42",
	})
	if err != nil {
		t.Fatalf("first CreateSubmission failed: %v", err)
	}

	secondPath, err := driveRef.CreateSubmission(SubmissionRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-1",
		StudentID:    "student-42",
	})
	if err != nil {
		t.Fatalf("second CreateSubmission failed: %v", err)
	}

	if firstPath == secondPath {
		t.Fatalf("expected distinct submission paths, got %q", firstPath)
	}

	labSubmissionDir := filepath.Join(driveRoot, "submissions", "go101-lab01")
	entries, err := os.ReadDir(labSubmissionDir)
	if err != nil {
		t.Fatalf("read lab submission directory: %v", err)
	}

	jsonCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 2 {
		t.Fatalf("expected 2 submission envelopes, got %d", jsonCount)
	}
}

func writeExercisePackage(path, labID, version string) error {
	innerArchive, err := buildInnerArchive()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	manifest := []byte(`{"lab_id":"` + labID + `","version":"` + version + `"}`)
	if err := writeTarEntry(tarWriter, "manifest.json", manifest); err != nil {
		return err
	}
	return writeTarEntry(tarWriter, "exercise.tar.gz", innerArchive)
}

func buildInnerArchive() ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := writeTarEntry(tarWriter, "README.md", []byte("hello")); err != nil {
		return nil, err
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeTarEntry(tarWriter *tar.Writer, name string, body []byte) error {
	header := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(body)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	_, err := tarWriter.Write(body)
	return err
}

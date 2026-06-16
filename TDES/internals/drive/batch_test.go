package drive

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndDecryptSubmissions(t *testing.T) {
	driveRoot := t.TempDir()

	// 1. Generate keys
	recipientPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}
	pubKeyB64 := base64.StdEncoding.EncodeToString(recipientPrivateKey.PublicKey().Bytes())

	// 2. Prepare drive
	if err := PrepareDrive(driveRoot); err != nil {
		t.Fatalf("PrepareDrive failed: %v", err)
	}
	if err := PrepareDriveForSubmission(driveRoot, pubKeyB64); err != nil {
		t.Fatalf("PrepareDriveForSubmission failed: %v", err)
	}

	// Create exercise workspace
	exerciseDir := t.TempDir()
	manifestJSON := []byte(`{
		"lab_id": "math101",
		"version": "1.0.0",
		"submission": {
			"include_paths": ["solution.go"]
		}
	}`)
	if err := os.WriteFile(filepath.Join(exerciseDir, "manifest.json"), manifestJSON, 0644); err != nil {
		t.Fatalf("write mock manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(exerciseDir, "solution.go"), []byte("package math\n"), 0644); err != nil {
		t.Fatalf("write mock solution: %v", err)
	}

	// Create 2 valid submissions
	driveRef := &Drive{Path: driveRoot}
	_, err = driveRef.CreateSubmission(SubmissionRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-test",
		StudentID:    "student-01",
	})
	if err != nil {
		t.Fatalf("CreateSubmission 1 failed: %v", err)
	}

	_, err = driveRef.CreateSubmission(SubmissionRequest{
		ExercisePath: exerciseDir,
		OrgID:        "org-test",
		StudentID:    "student-02",
	})
	if err != nil {
		t.Fatalf("CreateSubmission 2 failed: %v", err)
	}

	// 3. Write a random non-envelope JSON file (should be ignored)
	ignoredJSON := []byte(`{"hello": "world"}`)
	ignoredPath := filepath.Join(driveRoot, "submissions", "math101", "other-metadata.json")
	if err := os.WriteFile(ignoredPath, ignoredJSON, 0644); err != nil {
		t.Fatalf("write ignored JSON: %v", err)
	}

	// 4. Write a corrupted envelope JSON file (should report error)
	corruptEnvelope := []byte(`{
		"schema_version": "v1",
		"ciphertext_b64": "invalid-base64-content!!"
	}`)
	corruptPath := filepath.Join(driveRoot, "submissions", "math101", "submission-corrupt.json")
	if err := os.WriteFile(corruptPath, corruptEnvelope, 0644); err != nil {
		t.Fatalf("write corrupt JSON: %v", err)
	}

	// 5. Run LoadAndDecryptSubmissions
	results, err := LoadAndDecryptSubmissions(driveRoot, recipientPrivateKey)
	if err != nil {
		t.Fatalf("LoadAndDecryptSubmissions failed: %v", err)
	}

	// We expect 3 results: 2 valid decrypted ones, and 1 corrupt envelope error.
	// The other-metadata.json is ignored entirely because it doesn't look like an envelope.
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	validCount := 0
	errorCount := 0

	for _, res := range results {
		if res.Error != nil {
			errorCount++
			if !filepath.HasPrefix(filepath.Base(res.EnvelopePath), "submission-corrupt.json") {
				t.Errorf("unexpected error on file %s: %v", res.EnvelopePath, res.Error)
			}
		} else {
			validCount++
			if len(res.PlaintextTar) == 0 {
				t.Errorf("expected non-empty plaintext tar for %s", res.EnvelopePath)
			}
		}
	}

	if validCount != 2 {
		t.Errorf("expected 2 valid decrypted submissions, got %d", validCount)
	}
	if errorCount != 1 {
		t.Errorf("expected 1 errored submission, got %d", errorCount)
	}
}

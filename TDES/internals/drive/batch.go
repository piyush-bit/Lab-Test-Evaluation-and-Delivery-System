package drive

import (
	"crypto/ecdh"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DecryptedSubmission holds the decryption outcome of a discovered submission envelope.
type DecryptedSubmission struct {
	EnvelopePath    string
	PlaintextTar    []byte
	PlaintextSHA256 string
	Error           error
}

// LoadAndDecryptSubmissions scans the submissions directory under a drive root,
// finds all .json envelopes, decrypts them using the recipient private key,
// and returns their plaintext bytes. If a file fails to decrypt, the error is captured
// per-submission rather than stopping execution.
func LoadAndDecryptSubmissions(drivePath string, privateKey *ecdh.PrivateKey) ([]DecryptedSubmission, error) {
	resolvedPath, err := filepath.Abs(drivePath)
	if err != nil {
		return nil, fmt.Errorf("resolve drive path: %w", err)
	}

	submissionsDir := filepath.Join(resolvedPath, submissionDirName)
	info, err := os.Stat(submissionsDir)
	if err != nil {
		return nil, fmt.Errorf("submissions directory %q not found on drive: %w", submissionsDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("submissions path %q is not a directory", submissionsDir)
	}

	var results []DecryptedSubmission

	err = filepath.Walk(submissionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only look for .json files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		// Skip manifest.json
		if info.Name() == submissionManifestFileName {
			return nil
		}

		// Read envelope file
		data, err := os.ReadFile(path)
		if err != nil {
			results = append(results, DecryptedSubmission{
				EnvelopePath: path,
				Error:        fmt.Errorf("read file: %w", err),
			})
			return nil
		}

		var envelope SubmissionEnvelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			// Not a valid JSON or not unmarshalable to SubmissionEnvelope
			results = append(results, DecryptedSubmission{
				EnvelopePath: path,
				Error:        fmt.Errorf("decode envelope JSON: %w", err),
			})
			return nil
		}

		// Basic sanity check to verify this is indeed a submission envelope
		if envelope.SchemaVersion == "" || envelope.CiphertextB64 == "" {
			// It might be some other JSON file; skip it
			return nil
		}

		plaintext, err := DecryptSubmissionArchive(envelope, privateKey)
		if err != nil {
			results = append(results, DecryptedSubmission{
				EnvelopePath: path,
				Error:        fmt.Errorf("decrypt archive: %w", err),
			})
			return nil
		}

		results = append(results, DecryptedSubmission{
			EnvelopePath:    path,
			PlaintextTar:    plaintext,
			PlaintextSHA256: envelope.PlaintextSHA256,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk submissions: %w", err)
	}

	return results, nil
}

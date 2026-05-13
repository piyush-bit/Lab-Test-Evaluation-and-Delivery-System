package drive

import (
	"archive/tar"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"euc2/internals/exercise"
)

const (
	submissionDirName              = "submissions"
	submissionManifestFileName     = "manifest.json"
	submissionManifestSchema       = "v1"
	submissionEnvelopeSchema       = "v1"
	submissionEncryptAlgorithm     = "x25519-aes256-gcm"
	defaultSubmissionFileMode      = 0644
	defaultSubmissionDirectoryMode = 0755
)

type SubmissionManifest struct {
	SchemaVersion         string `json:"schema_version"`
	EncryptAlg            string `json:"encrypt_alg"`
	RecipientPublicKeyB64 string `json:"recipient_public_key_b64"`
}

type SubmissionRequest struct {
	ExercisePath string
	OrgID        string
	StudentID    string
}

type SubmissionEnvelope struct {
	SchemaVersion         string `json:"schema_version"`
	ArchiveFormat         string `json:"archive_format"`
	EncryptAlg            string `json:"encrypt_alg"`
	RecipientPublicKeyB64 string `json:"recipient_public_key_b64"`
	EphemeralPublicKeyB64 string `json:"ephemeral_public_key_b64"`
	NonceB64              string `json:"nonce_b64"`
	PlaintextSHA256       string `json:"plaintext_sha256"`
	CiphertextSHA256      string `json:"ciphertext_sha256"`
	CiphertextB64         string `json:"ciphertext_b64"`
}

func PrepareDriveForSubmission(drivePath string, recipientPublicKeyB64 string) error {
	drive, err := resolveDrive(drivePath, true)
	if err != nil {
		return err
	}

	recipientPublicKeyB64 = strings.TrimSpace(recipientPublicKeyB64)
	if recipientPublicKeyB64 == "" {
		return fmt.Errorf("recipient public key is required")
	}
	if _, err := decodeRecipientPublicKey(recipientPublicKeyB64); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(drive.Path, submissionDirName), defaultSubmissionDirectoryMode); err != nil {
		return fmt.Errorf("create drive submissions directory: %w", err)
	}

	drive.Manifest.ActiveModules["delivery"] = true
	if err := drive.Manifest.WriteManifest(drive.Path); err != nil {
		return fmt.Errorf("write drive manifest: %w", err)
	}

	manifest := SubmissionManifest{
		SchemaVersion:         submissionManifestSchema,
		EncryptAlg:            submissionEncryptAlgorithm,
		RecipientPublicKeyB64: recipientPublicKeyB64,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode submission manifest: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(submissionManifestPath(drive.Path), data, defaultSubmissionFileMode); err != nil {
		return fmt.Errorf("write submission manifest: %w", err)
	}
	return nil
}

func (d *Drive) CreateSubmission(request SubmissionRequest) (string, error) {
	if d == nil {
		return "", fmt.Errorf("drive is required")
	}

	resolvedDrive, err := resolveDrive(d.Path, false)
	if err != nil {
		return "", err
	}
	d.Path = resolvedDrive.Path
	d.Manifest = resolvedDrive.Manifest

	submissionManifest, err := loadSubmissionManifest(d.Path)
	if err != nil {
		return "", err
	}

	exercisePath, err := resolveExercisePath(request.ExercisePath)
	if err != nil {
		return "", err
	}

	manifest, err := exercise.LoadManifest(exercisePath)
	if err != nil {
		return "", fmt.Errorf("load exercise manifest: %w", err)
	}
	if strings.TrimSpace(manifest.LabID) == "" {
		return "", fmt.Errorf("manifest is missing lab_id")
	}

	submissionArchive, err := os.CreateTemp("", "submission-*.tar")
	if err != nil {
		return "", fmt.Errorf("create temp submission archive: %w", err)
	}
	defer func() {
		submissionArchive.Close()
		os.Remove(submissionArchive.Name())
	}()

	exerciseRef := &exercise.Exercise{
		Path:     exercisePath,
		Manifest: manifest,
	}
	if err := exerciseRef.CreateSubmissionPackage(submissionArchive, request.OrgID, request.StudentID); err != nil {
		return "", fmt.Errorf("create submission package: %w", err)
	}

	archiveBytes, err := os.ReadFile(submissionArchive.Name())
	if err != nil {
		return "", fmt.Errorf("read submission package: %w", err)
	}

	envelope, err := encryptSubmissionArchive(*submissionManifest, archiveBytes)
	if err != nil {
		return "", err
	}

	envelopeBytes, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode submission envelope: %w", err)
	}
	envelopeBytes = append(envelopeBytes, '\n')

	destinationDir := filepath.Join(d.Path, submissionDirName, manifest.LabID)
	if err := os.MkdirAll(destinationDir, defaultSubmissionDirectoryMode); err != nil {
		return "", fmt.Errorf("create exercise submission directory: %w", err)
	}

	destinationPath, err := writeSubmissionEnvelope(destinationDir, envelopeBytes)
	if err != nil {
		return "", err
	}
	return destinationPath, nil
}

func writeSubmissionEnvelope(destinationDir string, envelopeBytes []byte) (string, error) {
	file, err := os.CreateTemp(destinationDir, "submission-*.json")
	if err != nil {
		return "", fmt.Errorf("create submission envelope file: %w", err)
	}

	path := file.Name()
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.Write(envelopeBytes); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write submission envelope: %w", err)
	}
	if err := file.Chmod(defaultSubmissionFileMode); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("set submission envelope permissions: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close submission envelope file: %w", err)
	}
	return path, nil
}

func resolveExercisePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve exercise path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("access exercise path %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("exercise path %q is not a directory", absPath)
	}
	return absPath, nil
}

func loadSubmissionManifest(driveRoot string) (*SubmissionManifest, error) {
	data, err := os.ReadFile(submissionManifestPath(driveRoot))
	if err != nil {
		return nil, fmt.Errorf("read submission manifest: %w", err)
	}

	var manifest SubmissionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode submission manifest: %w", err)
	}

	if strings.TrimSpace(manifest.SchemaVersion) != submissionManifestSchema {
		return nil, fmt.Errorf("unsupported submission manifest schema %q", manifest.SchemaVersion)
	}
	if strings.TrimSpace(manifest.EncryptAlg) != submissionEncryptAlgorithm {
		return nil, fmt.Errorf("unsupported submission encrypt algorithm %q", manifest.EncryptAlg)
	}
	if _, err := decodeRecipientPublicKey(manifest.RecipientPublicKeyB64); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func submissionManifestPath(driveRoot string) string {
	return filepath.Join(driveRoot, submissionDirName, submissionManifestFileName)
}

func decodeRecipientPublicKey(publicKeyB64 string) (*ecdh.PublicKey, error) {
	rawKey, err := base64.StdEncoding.DecodeString(strings.TrimSpace(publicKeyB64))
	if err != nil {
		return nil, fmt.Errorf("decode recipient public key: %w", err)
	}

	publicKey, err := ecdh.X25519().NewPublicKey(rawKey)
	if err != nil {
		return nil, fmt.Errorf("parse recipient public key: %w", err)
	}
	return publicKey, nil
}

func encryptSubmissionArchive(manifest SubmissionManifest, plaintext []byte) (SubmissionEnvelope, error) {
	recipientPublicKey, err := decodeRecipientPublicKey(manifest.RecipientPublicKeyB64)
	if err != nil {
		return SubmissionEnvelope{}, err
	}

	ephemeralPrivateKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return SubmissionEnvelope{}, fmt.Errorf("generate ephemeral key: %w", err)
	}

	sharedSecret, err := ephemeralPrivateKey.ECDH(recipientPublicKey)
	if err != nil {
		return SubmissionEnvelope{}, fmt.Errorf("derive shared secret: %w", err)
	}

	keyMaterial := sha256.Sum256(append(append(append([]byte("euc2-submission-envelope"), sharedSecret...), recipientPublicKey.Bytes()...), ephemeralPrivateKey.PublicKey().Bytes()...))
	block, err := aes.NewCipher(keyMaterial[:])
	if err != nil {
		return SubmissionEnvelope{}, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return SubmissionEnvelope{}, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return SubmissionEnvelope{}, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, []byte(submissionEnvelopeSchema))
	plaintextHash := sha256.Sum256(plaintext)
	ciphertextHash := sha256.Sum256(ciphertext)

	return SubmissionEnvelope{
		SchemaVersion:         submissionEnvelopeSchema,
		ArchiveFormat:         "tar",
		EncryptAlg:            manifest.EncryptAlg,
		RecipientPublicKeyB64: manifest.RecipientPublicKeyB64,
		EphemeralPublicKeyB64: base64.StdEncoding.EncodeToString(ephemeralPrivateKey.PublicKey().Bytes()),
		NonceB64:              base64.StdEncoding.EncodeToString(nonce),
		PlaintextSHA256:       hex.EncodeToString(plaintextHash[:]),
		CiphertextSHA256:      hex.EncodeToString(ciphertextHash[:]),
		CiphertextB64:         base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func DecryptSubmissionArchive(envelope SubmissionEnvelope, recipientPrivateKey *ecdh.PrivateKey) ([]byte, error) {
	ephemeralKeyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(envelope.EphemeralPublicKeyB64))
	if err != nil {
		return nil, fmt.Errorf("decode ephemeral public key: %w", err)
	}
	ephemeralPublicKey, err := ecdh.X25519().NewPublicKey(ephemeralKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse ephemeral public key: %w", err)
	}

	sharedSecret, err := recipientPrivateKey.ECDH(ephemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("derive shared secret: %w", err)
	}

	recipientPublicKey, err := decodeRecipientPublicKey(envelope.RecipientPublicKeyB64)
	if err != nil {
		return nil, err
	}
	keyMaterial := sha256.Sum256(append(append(append([]byte("euc2-submission-envelope"), sharedSecret...), recipientPublicKey.Bytes()...), ephemeralPublicKey.Bytes()...))
	block, err := aes.NewCipher(keyMaterial[:])
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(strings.TrimSpace(envelope.NonceB64))
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(envelope.CiphertextB64))
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(submissionEnvelopeSchema))
	if err != nil {
		return nil, fmt.Errorf("decrypt submission archive: %w", err)
	}

	plaintextHash := sha256.Sum256(plaintext)
	if envelope.PlaintextSHA256 != hex.EncodeToString(plaintextHash[:]) {
		return nil, fmt.Errorf("plaintext sha256 mismatch")
	}
	return plaintext, nil
}

func readTarEntries(path string) ([]string, map[string][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := tar.NewReader(file)
	var order []string
	entries := make(map[string][]byte)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, nil, err
		}
		name := filepath.ToSlash(header.Name)
		order = append(order, name)
		entries[name] = body
	}
	return order, entries, nil
}

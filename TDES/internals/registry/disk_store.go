package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const hashShardLength = 2

// DiskArtifactStore implements ArtifactStore using the local filesystem.
type DiskArtifactStore struct {
	root string
}

// NewDiskArtifactStore creates and initializes a new DiskArtifactStore.
func NewDiskArtifactStore(root string) (*DiskArtifactStore, error) {
	if root == "" {
		return nil, fmt.Errorf("artifact root is required")
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("create artifact root: %w", err)
	}
	return &DiskArtifactStore{root: root}, nil
}

// Put uploads an archive from localPath into the content-addressed filesystem.
func (s *DiskArtifactStore) Put(ctx context.Context, localPath string) (Artifact, error) {
	hashValue, sizeBytes, err := hashFile(localPath)
	if err != nil {
		return Artifact{}, err
	}

	objectKey := objectKeyForHash(hashValue)
	destPath := filepath.Join(s.root, filepath.FromSlash(objectKey))
	if err := copyIfAbsent(localPath, destPath); err != nil {
		return Artifact{}, fmt.Errorf("persist artifact: %w", err)
	}

	return Artifact{
		SHA256:    hashValue,
		ObjectKey: objectKey,
		SizeBytes: sizeBytes,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// Open retrieves a read stream for the physical archive matching the SHA-256 hash.
func (s *DiskArtifactStore) Open(ctx context.Context, sha string) (io.ReadCloser, Artifact, error) {
	objectKey := objectKeyForHash(sha)
	path := filepath.Join(s.root, filepath.FromSlash(objectKey))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, Artifact{}, ErrNotFound
		}
		return nil, Artifact{}, fmt.Errorf("stat artifact: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, Artifact{}, fmt.Errorf("open artifact: %w", err)
	}

	return file, Artifact{
		SHA256:    sha,
		ObjectKey: objectKey,
		SizeBytes: info.Size(),
		CreatedAt: info.ModTime().UTC(),
	}, nil
}

// Exists checks if an archive matching the SHA-256 hash is in the store.
func (s *DiskArtifactStore) Exists(ctx context.Context, sha string) (bool, error) {
	objectKey := objectKeyForHash(sha)
	path := filepath.Join(s.root, filepath.FromSlash(objectKey))

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("check artifact existence: %w", err)
}

// Size returns the size of the physical archive file in bytes.
func (s *DiskArtifactStore) Size(ctx context.Context, sha string) (int64, error) {
	objectKey := objectKeyForHash(sha)
	path := filepath.Join(s.root, filepath.FromSlash(objectKey))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("stat artifact size: %w", err)
	}
	return info.Size(), nil
}

// Delete deletes the physical archive file from disk.
func (s *DiskArtifactStore) Delete(ctx context.Context, sha string) error {
	objectKey := objectKeyForHash(sha)
	path := filepath.Join(s.root, filepath.FromSlash(objectKey))

	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete artifact file: %w", err)
	}
	return nil
}

func objectKeyForHash(sha string) string {
	shard := sha
	if len(shard) > hashShardLength {
		shard = shard[:hashShardLength]
	}
	return filepath.ToSlash(filepath.Join("sha256", shard, sha))
}

func hashFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, fmt.Errorf("hash file content: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func copyIfAbsent(srcPath, destPath string) error {
	if _, err := os.Stat(destPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	tempFile, err := os.CreateTemp(filepath.Dir(destPath), "artifact-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, srcFile); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		if _, statErr := os.Stat(destPath); statErr == nil {
			return nil
		}
		return err
	}
	return nil
}

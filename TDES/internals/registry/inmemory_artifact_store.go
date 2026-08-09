package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type storedArtifactData struct {
	artifact Artifact
	content  []byte
}

// InMemoryArtifactStore implements ArtifactStore entirely in-memory.
type InMemoryArtifactStore struct {
	mu        sync.RWMutex
	artifacts map[string]storedArtifactData
}

// NewInMemoryArtifactStore creates a new thread-safe InMemoryArtifactStore.
func NewInMemoryArtifactStore() ArtifactStore {
	return &InMemoryArtifactStore{
		artifacts: make(map[string]storedArtifactData),
	}
}

func (s *InMemoryArtifactStore) Put(ctx context.Context, localPath string) (Artifact, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return Artifact{}, fmt.Errorf("read local path for in-memory store: %w", err)
	}

	hash := sha256.Sum256(data)
	sha := hex.EncodeToString(hash[:])

	s.mu.Lock()
	defer s.mu.Unlock()

	objectKey := objectKeyForHash(sha)
	art := Artifact{
		SHA256:    sha,
		ObjectKey: objectKey,
		SizeBytes: int64(len(data)),
		CreatedAt: time.Now().UTC(),
	}

	if _, exists := s.artifacts[sha]; !exists {
		s.artifacts[sha] = storedArtifactData{
			artifact: art,
			content:  data,
		}
	}

	return art, nil
}

func (s *InMemoryArtifactStore) Open(ctx context.Context, sha string) (io.ReadCloser, Artifact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.artifacts[sha]
	if !ok {
		return nil, Artifact{}, ErrNotFound
	}

	reader := io.NopCloser(bytes.NewReader(item.content))
	return reader, item.artifact, nil
}

func (s *InMemoryArtifactStore) Exists(ctx context.Context, sha string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.artifacts[sha]
	return ok, nil
}

func (s *InMemoryArtifactStore) Size(ctx context.Context, sha string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.artifacts[sha]
	if !ok {
		return 0, ErrNotFound
	}
	return item.artifact.SizeBytes, nil
}

func (s *InMemoryArtifactStore) Delete(ctx context.Context, sha string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.artifacts[sha]; !ok {
		return ErrNotFound
	}
	delete(s.artifacts, sha)
	return nil
}

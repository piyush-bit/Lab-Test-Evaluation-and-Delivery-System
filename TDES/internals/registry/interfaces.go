package registry

import (
	"context"
	"errors"
	"io"
)

var (
	// ErrNotFound is returned when an artifact or exercise version is not found.
	ErrNotFound = errors.New("not found")

	// ErrExerciseVersionConflict is returned when attempting to register a version that already exists with different metadata.
	ErrExerciseVersionConflict = errors.New("exercise version already exists with different metadata")
)

// ArtifactStore abstracts raw binary storage operations.
type ArtifactStore interface {
	// Put streams a local file archive to the store, calculates its SHA-256 hash, and returns metadata.
	Put(ctx context.Context, localPath string) (Artifact, error)

	// Open retrieves a read stream for the artifact matching the given SHA-256 content hash.
	Open(ctx context.Context, sha string) (io.ReadCloser, Artifact, error)

	// Exists checks if the artifact with the given hash already exists in storage.
	Exists(ctx context.Context, sha string) (bool, error)

	// Size returns the physical size of the artifact in bytes.
	Size(ctx context.Context, sha string) (int64, error)

	// Delete removes the archive from storage.
	Delete(ctx context.Context, sha string) error
}

// Repository abstracts registry metadata and database operations.
type Repository interface {
	// UpsertArtifact registers/updates artifact metadata in the index.
	UpsertArtifact(ctx context.Context, artifact Artifact) error

	// GetArtifact retrieves artifact metadata by SHA-256 hash. Returns ErrNotFound if missing.
	GetArtifact(ctx context.Context, sha string) (Artifact, error)

	// CreateOrGetExerciseVersion registers a new version or gets the existing one.
	CreateOrGetExerciseVersion(ctx context.Context, version ExerciseVersion) (ExerciseVersion, bool, error)

	// GetExerciseVersion retrieves exercise version metadata. Returns ErrNotFound if missing.
	GetExerciseVersion(ctx context.Context, orgID, exerciseID, version string) (ExerciseVersion, error)

	// ListExercises lists all registered exercise versions, optionally filtered by orgID and status (if not empty).
	ListExercises(ctx context.Context, orgID, status string) ([]ExerciseVersion, error)

	// ListExerciseVersions lists all versions of a specific exercise.
	ListExerciseVersions(ctx context.Context, orgID, exerciseID string) ([]ExerciseVersion, error)

	// UpdateExerciseStatus updates the status (e.g., "draft", "published", "retired") of a version.
	UpdateExerciseStatus(ctx context.Context, orgID, exerciseID, version, status string) error

	// DeleteExerciseVersion removes an exercise version metadata entry.
	DeleteExerciseVersion(ctx context.Context, orgID, exerciseID, version string) error
}

package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"TDES/internals/exercise"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteRepository implements Repository on top of SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLiteRepository and initializes its database schemas.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path is required")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.initSchema(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return repo, nil
}

// Close closes the database connection.
func (r *SQLiteRepository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *SQLiteRepository) initSchema(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS artifacts (
	sha256 TEXT PRIMARY KEY,
	object_key TEXT NOT NULL UNIQUE,
	size_bytes INTEGER NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS exercise_versions (
	id TEXT PRIMARY KEY,
	org_id TEXT NOT NULL,
	exercise_id TEXT NOT NULL,
	version TEXT NOT NULL,
	title TEXT NOT NULL,
	language TEXT NOT NULL,
	status TEXT NOT NULL,
	manifest_json TEXT NOT NULL,
	public_artifact_sha256 TEXT NOT NULL,
	private_artifact_sha256 TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	UNIQUE (org_id, exercise_id, version),
	FOREIGN KEY (public_artifact_sha256) REFERENCES artifacts(sha256),
	FOREIGN KEY (private_artifact_sha256) REFERENCES artifacts(sha256)
);
`
	if _, err := r.db.ExecContext(ctx, schema); err != nil {
		return err
	}
	return nil
}

// UpsertArtifact registers artifact metadata in the index.
func (r *SQLiteRepository) UpsertArtifact(ctx context.Context, artifact Artifact) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO artifacts (sha256, object_key, size_bytes, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(sha256) DO NOTHING`,
		artifact.SHA256,
		artifact.ObjectKey,
		artifact.SizeBytes,
		now,
	)
	if err != nil {
		return fmt.Errorf("db upsert artifact: %w", err)
	}
	return nil
}

// GetArtifact retrieves artifact metadata by SHA-256 hash. Returns ErrNotFound if missing.
func (r *SQLiteRepository) GetArtifact(ctx context.Context, sha string) (Artifact, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT sha256, object_key, size_bytes, created_at FROM artifacts WHERE sha256 = ?`,
		sha,
	)

	var (
		artifact  Artifact
		createdAt string
	)
	if err := row.Scan(
		&artifact.SHA256,
		&artifact.ObjectKey,
		&artifact.SizeBytes,
		&createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artifact{}, ErrNotFound
		}
		return Artifact{}, fmt.Errorf("db scan artifact: %w", err)
	}

	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Artifact{}, fmt.Errorf("db parse created_at: %w", err)
	}
	artifact.CreatedAt = parsedCreated

	return artifact, nil
}

// CreateOrGetExerciseVersion registers a new version or gets the existing one.
func (r *SQLiteRepository) CreateOrGetExerciseVersion(ctx context.Context, exerciseVersion ExerciseVersion) (ExerciseVersion, bool, error) {
	manifestBytes, err := json.Marshal(exerciseVersion.ManifestSnapshot)
	if err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("db encode manifest: %w", err)
	}

	now := time.Now().UTC()
	if exerciseVersion.ID == "" {
		exerciseVersion.ID = uuid.NewString()
	}
	exerciseVersion.CreatedAt = now
	exerciseVersion.UpdatedAt = now

	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO exercise_versions (
			id, org_id, exercise_id, version, title, language, status, manifest_json,
			public_artifact_sha256, private_artifact_sha256, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		exerciseVersion.ID,
		exerciseVersion.OrgID,
		exerciseVersion.ExerciseID,
		exerciseVersion.Version,
		exerciseVersion.Title,
		exerciseVersion.Language,
		exerciseVersion.Status,
		string(manifestBytes),
		exerciseVersion.PublicArtifactSHA,
		exerciseVersion.PrivateArtifactSHA,
		exerciseVersion.CreatedAt.Format(time.RFC3339Nano),
		exerciseVersion.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err == nil {
		affected, _ := result.RowsAffected()
		return exerciseVersion, affected == 1, nil
	}

	// Uniqueness constraint hit, attempt to fetch the existing record
	current, findErr := r.GetExerciseVersion(ctx, exerciseVersion.OrgID, exerciseVersion.ExerciseID, exerciseVersion.Version)
	if findErr != nil {
		return ExerciseVersion{}, false, fmt.Errorf("create exercise version duplicate check: %w", err)
	}

	if !sameExerciseVersion(current, exerciseVersion) {
		return ExerciseVersion{}, false, ErrExerciseVersionConflict
	}
	return current, false, nil
}

// GetExerciseVersion retrieves exercise version metadata. Returns ErrNotFound if missing.
func (r *SQLiteRepository) GetExerciseVersion(ctx context.Context, orgID, exerciseID, version string) (ExerciseVersion, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, org_id, exercise_id, version, title, language, status, manifest_json,
		        public_artifact_sha256, private_artifact_sha256, created_at, updated_at
		   FROM exercise_versions
		  WHERE org_id = ? AND exercise_id = ? AND version = ?`,
		orgID,
		exerciseID,
		version,
	)
	return r.scanVersion(row)
}

// ListExercises lists all registered exercise versions, optionally filtered by orgID and status (if not empty).
func (r *SQLiteRepository) ListExercises(ctx context.Context, orgID, status string) ([]ExerciseVersion, error) {
	query := `
		SELECT id, org_id, exercise_id, version, title, language, status, manifest_json,
		        public_artifact_sha256, private_artifact_sha256, created_at, updated_at
		   FROM exercise_versions
		  WHERE (? = '' OR org_id = ?) AND (? = '' OR status = ?)
		  ORDER BY org_id, exercise_id, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, orgID, status, status)
	if err != nil {
		return nil, fmt.Errorf("db list exercises: %w", err)
	}
	defer rows.Close()

	var versions []ExerciseVersion
	for rows.Next() {
		ev, err := r.scanVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, ev)
	}
	return versions, nil
}

// ListExerciseVersions lists all versions of a specific exercise.
func (r *SQLiteRepository) ListExerciseVersions(ctx context.Context, orgID, exerciseID string) ([]ExerciseVersion, error) {
	query := `
		SELECT id, org_id, exercise_id, version, title, language, status, manifest_json,
		        public_artifact_sha256, private_artifact_sha256, created_at, updated_at
		   FROM exercise_versions
		  WHERE org_id = ? AND exercise_id = ?
		  ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("db list exercise versions: %w", err)
	}
	defer rows.Close()

	var versions []ExerciseVersion
	for rows.Next() {
		ev, err := r.scanVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, ev)
	}
	return versions, nil
}

// UpdateExerciseStatus updates the status of a version. Returns ErrNotFound if version missing.
func (r *SQLiteRepository) UpdateExerciseStatus(ctx context.Context, orgID, exerciseID, version, status string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE exercise_versions SET status = ?, updated_at = ?
		 WHERE org_id = ? AND exercise_id = ? AND version = ?`,
		status,
		now,
		orgID,
		exerciseID,
		version,
	)
	if err != nil {
		return fmt.Errorf("db update status: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteExerciseVersion removes an exercise version metadata entry. Returns ErrNotFound if version missing.
func (r *SQLiteRepository) DeleteExerciseVersion(ctx context.Context, orgID, exerciseID, version string) error {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM exercise_versions
		 WHERE org_id = ? AND exercise_id = ? AND version = ?`,
		orgID,
		exerciseID,
		version,
	)
	if err != nil {
		return fmt.Errorf("db delete exercise version: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// scannable represents anything that scanVersion can scan (sql.Row or sql.Rows)
type scannable interface {
	Scan(dest ...any) error
}

func (r *SQLiteRepository) scanVersion(row scannable) (ExerciseVersion, error) {
	var (
		ev           ExerciseVersion
		manifestJSON string
		createdAt    string
		updatedAt    string
	)
	if err := row.Scan(
		&ev.ID,
		&ev.OrgID,
		&ev.ExerciseID,
		&ev.Version,
		&ev.Title,
		&ev.Language,
		&ev.Status,
		&manifestJSON,
		&ev.PublicArtifactSHA,
		&ev.PrivateArtifactSHA,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExerciseVersion{}, ErrNotFound
		}
		return ExerciseVersion{}, fmt.Errorf("scan version error: %w", err)
	}

	var manifest exercise.ExerciseManifest
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		return ExerciseVersion{}, fmt.Errorf("db unmarshal manifest: %w", err)
	}
	ev.ManifestSnapshot = &manifest

	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return ExerciseVersion{}, fmt.Errorf("db parse created_at: %w", err)
	}
	ev.CreatedAt = parsedCreated

	parsedUpdated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return ExerciseVersion{}, fmt.Errorf("db parse updated_at: %w", err)
	}
	ev.UpdatedAt = parsedUpdated

	return ev, nil
}

func sameExerciseVersion(current, candidate ExerciseVersion) bool {
	currBytes, _ := json.Marshal(current.ManifestSnapshot)
	candBytes, _ := json.Marshal(candidate.ManifestSnapshot)

	return current.OrgID == candidate.OrgID &&
		current.ExerciseID == candidate.ExerciseID &&
		current.Version == candidate.Version &&
		current.Title == candidate.Title &&
		current.Language == candidate.Language &&
		current.Status == candidate.Status &&
		current.PublicArtifactSHA == candidate.PublicArtifactSHA &&
		current.PrivateArtifactSHA == candidate.PrivateArtifactSHA &&
		string(currBytes) == string(candBytes)
}

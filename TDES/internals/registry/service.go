package registry

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Service coordinates Repository and ArtifactStore logic.
type Service struct {
	repo      Repository
	artifacts ArtifactStore
}

// NewService instantiates a new Service.
func NewService(repo Repository, artifacts ArtifactStore) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if artifacts == nil {
		return nil, fmt.Errorf("artifact store is required")
	}
	return &Service{repo: repo, artifacts: artifacts}, nil
}

// Publish registers a new version of an exercise and stores its public/private packages.
func (s *Service) Publish(ctx context.Context, request PublishRequest) (ExerciseVersion, bool, error) {
	if strings.TrimSpace(request.OrgID) == "" {
		return ExerciseVersion{}, false, fmt.Errorf("org_id is required")
	}
	if strings.TrimSpace(request.PublicArtifactPath) == "" {
		return ExerciseVersion{}, false, fmt.Errorf("public artifact path is required")
	}
	if strings.TrimSpace(request.PrivateArtifactPath) == "" {
		return ExerciseVersion{}, false, fmt.Errorf("private artifact path is required")
	}

	manifest, err := ReadManifestFromPackage(request.PublicArtifactPath)
	if err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("read public manifest: %w", err)
	}

	exerciseID := strings.TrimSpace(request.ExerciseID)
	if exerciseID == "" {
		exerciseID = strings.TrimSpace(manifest.LabID)
	}
	if exerciseID == "" {
		return ExerciseVersion{}, false, fmt.Errorf("exercise_id is required")
	}

	version := strings.TrimSpace(request.Version)
	if version == "" {
		version = strings.TrimSpace(manifest.Version)
	}
	if version == "" {
		return ExerciseVersion{}, false, fmt.Errorf("version is required")
	}

	if strings.TrimSpace(manifest.LabID) != "" && manifest.LabID != exerciseID {
		return ExerciseVersion{}, false, fmt.Errorf("manifest lab_id %q does not match exercise_id %q", manifest.LabID, exerciseID)
	}
	if strings.TrimSpace(manifest.Version) != "" && manifest.Version != version {
		return ExerciseVersion{}, false, fmt.Errorf("manifest version %q does not match version %q", manifest.Version, version)
	}

	publicArtifact, err := s.artifacts.Put(ctx, request.PublicArtifactPath)
	if err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("store public artifact: %w", err)
	}
	if err := s.repo.UpsertArtifact(ctx, publicArtifact); err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("db save public artifact: %w", err)
	}

	privateArtifact, err := s.artifacts.Put(ctx, request.PrivateArtifactPath)
	if err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("store private artifact: %w", err)
	}
	if err := s.repo.UpsertArtifact(ctx, privateArtifact); err != nil {
		return ExerciseVersion{}, false, fmt.Errorf("db save private artifact: %w", err)
	}

	status := strings.TrimSpace(request.Status)
	if status == "" {
		status = "published"
	}

	exerciseVersion := ExerciseVersion{
		OrgID:              strings.TrimSpace(request.OrgID),
		ExerciseID:         exerciseID,
		Version:            version,
		Title:              strings.TrimSpace(manifest.Title),
		Language:           strings.TrimSpace(manifest.Language),
		Status:             status,
		ManifestSnapshot:   manifest,
		PublicArtifactSHA:  publicArtifact.SHA256,
		PrivateArtifactSHA: privateArtifact.SHA256,
	}

	stored, created, err := s.repo.CreateOrGetExerciseVersion(ctx, exerciseVersion)
	if err != nil {
		return ExerciseVersion{}, false, err
	}

	return stored, created, nil
}

// GetExerciseVersion retrieves metadata for a specific exercise version.
func (s *Service) GetExerciseVersion(ctx context.Context, orgID, exerciseID, version string) (ExerciseVersion, error) {
	return s.repo.GetExerciseVersion(ctx, orgID, exerciseID, version)
}

// ListExercises lists all registered exercises, optionally filtered by orgID and status.
func (s *Service) ListExercises(ctx context.Context, orgID, status string) ([]ExerciseVersion, error) {
	return s.repo.ListExercises(ctx, orgID, status)
}

// ListExerciseVersions lists all versions of a specific exercise.
func (s *Service) ListExerciseVersions(ctx context.Context, orgID, exerciseID string) ([]ExerciseVersion, error) {
	return s.repo.ListExerciseVersions(ctx, orgID, exerciseID)
}

// UpdateExerciseStatus updates the status (e.g., "draft", "published", "retired") of a version.
func (s *Service) UpdateExerciseStatus(ctx context.Context, orgID, exerciseID, version, status string) error {
	return s.repo.UpdateExerciseStatus(ctx, orgID, exerciseID, version, status)
}

// DeleteExerciseVersion removes an exercise version metadata record.
func (s *Service) DeleteExerciseVersion(ctx context.Context, orgID, exerciseID, version string) error {
	return s.repo.DeleteExerciseVersion(ctx, orgID, exerciseID, version)
}

// ArtifactHandle wraps a physical artifact stream with its metadata.
type ArtifactHandle struct {
	File     io.ReadCloser
	Artifact Artifact
}

// OpenArtifact retrieves a file stream and metadata for the given SHA-256 hash.
func (s *Service) OpenArtifact(ctx context.Context, sha string) (*ArtifactHandle, error) {
	file, artifact, err := s.artifacts.Open(ctx, strings.TrimSpace(sha))
	if err != nil {
		return nil, err
	}
	return &ArtifactHandle{File: file, Artifact: artifact}, nil
}

// SaveEvaluation stores a student submission evaluation result in the database.
func (s *Service) SaveEvaluation(ctx context.Context, eval SubmissionEvaluation) error {
	return s.repo.SaveEvaluation(ctx, eval)
}

// ListSubmissions retrieves student submission evaluations from the repository.
func (s *Service) ListSubmissions(ctx context.Context, orgID, labID string) ([]SubmissionEvaluation, error) {
	return s.repo.ListSubmissions(ctx, orgID, labID)
}

// GetStudentCredential retrieves a student's credential by orgID and studentID.
func (s *Service) GetStudentCredential(ctx context.Context, orgID, studentID string) (StudentCredential, error) {
	return s.repo.GetStudentCredential(ctx, orgID, studentID)
}

// SaveStudentCredential saves a student's credential (e.g. updating the pin_hash or creating a roster entry).
func (s *Service) SaveStudentCredential(ctx context.Context, cred StudentCredential) error {
	return s.repo.SaveStudentCredential(ctx, cred)
}

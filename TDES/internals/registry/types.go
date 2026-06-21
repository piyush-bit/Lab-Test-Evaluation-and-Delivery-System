package registry

import (
	"time"

	"TDES/internals/exercise"
)

// Artifact represents the metadata of a stored binary archive.
type Artifact struct {
	SHA256    string    `json:"sha256"`
	ObjectKey string    `json:"object_key"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

// ExerciseVersion represents an exercise entry in the registry db.
type ExerciseVersion struct {
	ID                 string                     `json:"id"`
	OrgID              string                     `json:"org_id"`
	ExerciseID         string                     `json:"exercise_id"`
	Version            string                     `json:"version"`
	Title              string                     `json:"title"`
	Language           string                     `json:"language"`
	Status             string                     `json:"status"`
	ManifestSnapshot   *exercise.ExerciseManifest `json:"manifest_snapshot"`
	PublicArtifactSHA  string                     `json:"public_artifact_sha256"`
	PrivateArtifactSHA string                     `json:"private_artifact_sha256"`
	CreatedAt          time.Time                  `json:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at"`
}

// PublishRequest holds data for registering a new exercise.
type PublishRequest struct {
	OrgID               string
	ExerciseID          string
	Version             string
	Status              string
	PublicArtifactPath  string
	PrivateArtifactPath string
}

// SubmissionEvaluation represents a student submission evaluation result stored in the database.
type SubmissionEvaluation struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	StudentID    string    `json:"student_id"`
	LabID        string    `json:"lab_id"`
	Version      string    `json:"version"`
	Status       string    `json:"status"`
	EarnedPoints int       `json:"earned_points"`
	MaxPoints    int       `json:"max_points"`
	ResultsJSON  string    `json:"results_json"`
	CreatedAt    time.Time `json:"created_at"`
}

// StudentCredential represents a student's private credentials stored in the database.
type StudentCredential struct {
	OrgID     string    `json:"org_id"`
	StudentID string    `json:"student_id"`
	PinHash   string    `json:"pin_hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

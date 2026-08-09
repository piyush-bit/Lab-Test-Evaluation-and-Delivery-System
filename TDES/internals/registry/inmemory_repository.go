package registry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryRepository implements Repository entirely in-memory using thread-safe maps.
type InMemoryRepository struct {
	mu                 sync.RWMutex
	artifacts          map[string]Artifact
	exerciseVersions   map[string]ExerciseVersion
	submissions        map[string]SubmissionEvaluation
	studentCredentials map[string]StudentCredential
	closed             bool
}

// NewInMemoryRepository creates a new thread-safe InMemoryRepository.
func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		artifacts:          make(map[string]Artifact),
		exerciseVersions:   make(map[string]ExerciseVersion),
		submissions:        make(map[string]SubmissionEvaluation),
		studentCredentials: make(map[string]StudentCredential),
	}
}

func (r *InMemoryRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	return nil
}

func (r *InMemoryRepository) UpsertArtifact(ctx context.Context, artifact Artifact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("repository is closed")
	}

	if artifact.CreatedAt.IsZero() {
		artifact.CreatedAt = time.Now().UTC()
	}
	if _, exists := r.artifacts[artifact.SHA256]; !exists {
		r.artifacts[artifact.SHA256] = artifact
	}
	return nil
}

func (r *InMemoryRepository) GetArtifact(ctx context.Context, sha string) (Artifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return Artifact{}, fmt.Errorf("repository is closed")
	}

	art, ok := r.artifacts[sha]
	if !ok {
		return Artifact{}, ErrNotFound
	}
	return art, nil
}

func (r *InMemoryRepository) CreateOrGetExerciseVersion(ctx context.Context, ev ExerciseVersion) (ExerciseVersion, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ExerciseVersion{}, false, fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s:%s", ev.OrgID, ev.ExerciseID, ev.Version)
	existing, found := r.exerciseVersions[key]
	if found {
		if !sameExerciseVersion(existing, ev) {
			return ExerciseVersion{}, false, ErrExerciseVersionConflict
		}
		return existing, false, nil
	}

	now := time.Now().UTC()
	if ev.ID == "" {
		ev.ID = uuid.NewString()
	}
	ev.CreatedAt = now
	ev.UpdatedAt = now

	r.exerciseVersions[key] = ev
	return ev, true, nil
}

func (r *InMemoryRepository) GetExerciseVersion(ctx context.Context, orgID, exerciseID, version string) (ExerciseVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return ExerciseVersion{}, fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s:%s", orgID, exerciseID, version)
	ev, ok := r.exerciseVersions[key]
	if !ok {
		return ExerciseVersion{}, ErrNotFound
	}
	return ev, nil
}

func (r *InMemoryRepository) ListExercises(ctx context.Context, orgID, status, search string) ([]ExerciseVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return nil, fmt.Errorf("repository is closed")
	}

	searchLower := strings.ToLower(search)
	var list []ExerciseVersion

	for _, ev := range r.exerciseVersions {
		if orgID != "" && ev.OrgID != orgID {
			continue
		}
		if status != "" && ev.Status != status {
			continue
		}
		if searchLower != "" {
			titleLower := strings.ToLower(ev.Title)
			exIDLower := strings.ToLower(ev.ExerciseID)
			langLower := strings.ToLower(ev.Language)
			if !strings.Contains(titleLower, searchLower) &&
				!strings.Contains(exIDLower, searchLower) &&
				!strings.Contains(langLower, searchLower) {
				continue
			}
		}
		list = append(list, ev)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].OrgID != list[j].OrgID {
			return list[i].OrgID < list[j].OrgID
		}
		if list[i].ExerciseID != list[j].ExerciseID {
			return list[i].ExerciseID < list[j].ExerciseID
		}
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	return list, nil
}

func (r *InMemoryRepository) ListExerciseVersions(ctx context.Context, orgID, exerciseID string) ([]ExerciseVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return nil, fmt.Errorf("repository is closed")
	}

	var list []ExerciseVersion
	for _, ev := range r.exerciseVersions {
		if ev.OrgID == orgID && ev.ExerciseID == exerciseID {
			list = append(list, ev)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	return list, nil
}

func (r *InMemoryRepository) UpdateExerciseStatus(ctx context.Context, orgID, exerciseID, version, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s:%s", orgID, exerciseID, version)
	ev, ok := r.exerciseVersions[key]
	if !ok {
		return ErrNotFound
	}

	ev.Status = status
	ev.UpdatedAt = time.Now().UTC()
	r.exerciseVersions[key] = ev
	return nil
}

func (r *InMemoryRepository) DeleteExerciseVersion(ctx context.Context, orgID, exerciseID, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s:%s", orgID, exerciseID, version)
	if _, ok := r.exerciseVersions[key]; !ok {
		return ErrNotFound
	}

	delete(r.exerciseVersions, key)
	return nil
}

func (r *InMemoryRepository) SaveEvaluation(ctx context.Context, eval SubmissionEvaluation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("repository is closed")
	}

	if eval.ID == "" {
		eval.ID = uuid.NewString()
	}
	if eval.CreatedAt.IsZero() {
		eval.CreatedAt = time.Now().UTC()
	}
	r.submissions[eval.ID] = eval
	return nil
}

func (r *InMemoryRepository) ListSubmissions(ctx context.Context, orgID, labID string) ([]SubmissionEvaluation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return nil, fmt.Errorf("repository is closed")
	}

	var list []SubmissionEvaluation
	for _, sub := range r.submissions {
		if orgID != "" && sub.OrgID != orgID {
			continue
		}
		if labID != "" && sub.LabID != labID {
			continue
		}
		list = append(list, sub)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	return list, nil
}

func (r *InMemoryRepository) GetStudentCredential(ctx context.Context, orgID, studentID string) (StudentCredential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return StudentCredential{}, fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s", orgID, studentID)
	cred, ok := r.studentCredentials[key]
	if !ok {
		return StudentCredential{}, ErrNotFound
	}
	return cred, nil
}

func (r *InMemoryRepository) SaveStudentCredential(ctx context.Context, cred StudentCredential) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("repository is closed")
	}

	key := fmt.Sprintf("%s:%s", cred.OrgID, cred.StudentID)
	now := time.Now().UTC()

	existing, found := r.studentCredentials[key]
	if found {
		existing.PinHash = cred.PinHash
		existing.UpdatedAt = now
		r.studentCredentials[key] = existing
	} else {
		cred.CreatedAt = now
		cred.UpdatedAt = now
		r.studentCredentials[key] = cred
	}
	return nil
}

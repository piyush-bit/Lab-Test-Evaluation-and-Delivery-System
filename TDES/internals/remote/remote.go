package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"TDES/internals/exercise"
)

const defaultTimeout = 30 * time.Second

type Remote struct {
	BaseURL     string
	Client      *http.Client
	PackageFunc func(path string) (string, string, error)
}

type ExerciseVersion struct {
	OrgID             string `json:"org_id"`
	ExerciseID        string `json:"exercise_id"`
	Version           string `json:"version"`
	PublicArtifactSHA string `json:"public_artifact_sha256"`
}

func NewRemote(baseURL string) *Remote {
	return &Remote{
		BaseURL:     baseURL,
		Client:      &http.Client{Timeout: defaultTimeout},
		PackageFunc: exercise.PackageExercise,
	}
}

func (r *Remote) baseURL() (string, error) {
	if r == nil {
		return "", fmt.Errorf("remote is required")
	}

	baseURL := strings.TrimSpace(r.BaseURL)
	if baseURL == "" {
		return "", fmt.Errorf("remote base url is required")
	}
	return strings.TrimRight(baseURL, "/"), nil
}

func (r *Remote) httpClient() *http.Client {
	if r != nil && r.Client != nil {
		return r.Client
	}
	return &http.Client{Timeout: defaultTimeout}
}

func (r *Remote) exerciseVersionURL(orgID string, exerciseID string, version string) (string, error) {
	baseURL, err := r.baseURL()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(orgID) == "" {
		return "", fmt.Errorf("org id is required")
	}
	if strings.TrimSpace(exerciseID) == "" {
		return "", fmt.Errorf("exercise id is required")
	}
	if strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("version is required")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse remote base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, "v1", "exercises", url.PathEscape(orgID), url.PathEscape(exerciseID), "versions", url.PathEscape(version))
	return parsedURL.String(), nil
}

func (r *Remote) artifactURL(sha string) (string, error) {
	baseURL, err := r.baseURL()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(sha) == "" {
		return "", fmt.Errorf("artifact sha is required")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse remote base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, "v1", "artifacts", url.PathEscape(strings.TrimSpace(sha)))
	return parsedURL.String(), nil
}

func (r *Remote) submissionURL() (string, error) {
	baseURL, err := r.baseURL()
	if err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse remote base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, "v1", "submissions")
	return parsedURL.String(), nil
}

func (r *Remote) publishURL() (string, error) {
	baseURL, err := r.baseURL()
	if err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse remote base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, "v1", "exercises", "publish")
	return parsedURL.String(), nil
}

func decodeJSONResponse[T any](response *http.Response, destination *T) error {
	if response == nil {
		return fmt.Errorf("response is required")
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = response.Status
		}
		return fmt.Errorf("remote request failed: %s", message)
	}

	if destination == nil {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

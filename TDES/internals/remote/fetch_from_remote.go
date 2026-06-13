package remote

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	exercisestore "TDES/internals/exercise_store"
)

func (r *Remote) FetchPrivateFromRemote(id string, version string, orgID string, bearerToken string) (io.ReadCloser, error) {
	baseURL, err := r.baseURL()
	if err != nil {
		return nil, err
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse remote base url: %w", err)
	}
	parsedURL.Path = path.Join(parsedURL.Path, "v1", "exercises", url.PathEscape(orgID), url.PathEscape(id), "versions", url.PathEscape(version), "download")
	q := parsedURL.Query()
	q.Set("type", "private")
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := r.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("download private artifact: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		resp.Body.Close()
		return nil, fmt.Errorf("download private artifact: remote request failed: %s", resp.Status)
	}

	return resp.Body, nil
}

func (r *Remote) FetchFromRemote(id string, version string, orgID string) error {
	metadataURL, err := r.exerciseVersionURL(orgID, id, version)
	if err != nil {
		return err
	}

	metadataResponse, err := r.httpClient().Get(metadataURL)
	if err != nil {
		return fmt.Errorf("fetch exercise metadata: %w", err)
	}

	var exerciseVersion ExerciseVersion
	if err := decodeJSONResponse(metadataResponse, &exerciseVersion); err != nil {
		return fmt.Errorf("fetch exercise metadata: %w", err)
	}

	artifactURL, err := r.artifactURL(exerciseVersion.PublicArtifactSHA)
	if err != nil {
		return err
	}

	artifactResponse, err := r.httpClient().Get(artifactURL)
	if err != nil {
		return fmt.Errorf("download public artifact: %w", err)
	}
	defer artifactResponse.Body.Close()

	if artifactResponse.StatusCode < http.StatusOK || artifactResponse.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download public artifact: remote request failed: %s", artifactResponse.Status)
	}

	tempFile, err := os.CreateTemp("", "remote-exercise-*.tar")
	if err != nil {
		return fmt.Errorf("create temp package: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	if _, err := tempFile.ReadFrom(artifactResponse.Body); err != nil {
		return fmt.Errorf("download public artifact: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp package: %w", err)
	}

	if err := exercisestore.SavePackage(exercisestore.GetPublicCacheDir(), tempPath); err != nil {
		return fmt.Errorf("save remote package to local cache: %w", err)
	}
	return nil
}

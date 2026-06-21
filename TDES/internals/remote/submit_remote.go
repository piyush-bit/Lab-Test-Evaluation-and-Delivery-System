package remote

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"TDES/internals/exercise"
)

const BearerTokenEnvVar = "EUC2_REMOTE_BEARER_TOKEN"

type SubmitRequest struct {
	ExercisePath string
	OrgID        string
	StudentID    string
	BearerToken  string
	Pin          string
	NewPin       string
}

func (r *Remote) SubmitRemote(request SubmitRequest) (string, error) {
	exercisePath, manifest, err := loadExerciseForSubmission(request.ExercisePath)
	if err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp("", "submission-*.tar")
	if err != nil {
		return "", fmt.Errorf("create temp submission archive: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempPath)
	}()

	exerciseRef := &exercise.Exercise{
		Path:     exercisePath,
		Manifest: manifest,
	}
	if err := exerciseRef.CreateSubmissionPackage(tempFile, request.OrgID, request.StudentID); err != nil {
		return "", fmt.Errorf("create submission package: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close submission package: %w", err)
	}

	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	part, err := writer.CreateFormFile("submission_package", filepath.Base(tempPath))
	if err != nil {
		return "", fmt.Errorf("create submission form file: %w", err)
	}

	submissionPackage, err := os.Open(tempPath)
	if err != nil {
		return "", fmt.Errorf("open submission package: %w", err)
	}
	if _, err := io.Copy(part, submissionPackage); err != nil {
		submissionPackage.Close()
		return "", fmt.Errorf("copy submission package: %w", err)
	}
	if err := submissionPackage.Close(); err != nil {
		return "", fmt.Errorf("close submission package: %w", err)
	}

	if request.Pin != "" {
		if err := writer.WriteField("pin", request.Pin); err != nil {
			return "", fmt.Errorf("write pin field: %w", err)
		}
	}
	if request.NewPin != "" {
		if err := writer.WriteField("new_pin", request.NewPin); err != nil {
			return "", fmt.Errorf("write new_pin field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finalize submission form: %w", err)
	}

	submissionURL, err := r.submissionURL()
	if err != nil {
		return "", err
	}

	httpRequest, err := http.NewRequest(http.MethodPost, submissionURL, bodyBuffer)
	if err != nil {
		return "", fmt.Errorf("create submission request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())

	bearerToken := resolveBearerToken(request.BearerToken)
	if bearerToken == "" {
		return "", fmt.Errorf("bearer token is required via %s", BearerTokenEnvVar)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+bearerToken)

	response, err := r.httpClient().Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("submit submission package: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8192))
	if err != nil {
		return "", fmt.Errorf("read submission response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(responseBody))
		if message == "" {
			message = response.Status
		}
		return "", fmt.Errorf("submit submission package: %s", message)
	}

	message := strings.TrimSpace(string(responseBody))
	if message == "" {
		message = response.Status
	}
	return message, nil
}

func loadExerciseForSubmission(path string) (string, *exercise.ExerciseManifest, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "."
	}

	exercisePath, err := filepath.Abs(path)
	if err != nil {
		return "", nil, fmt.Errorf("resolve exercise path: %w", err)
	}

	info, err := os.Stat(exercisePath)
	if err != nil {
		return "", nil, fmt.Errorf("access exercise path %q: %w", exercisePath, err)
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("exercise path %q is not a directory", exercisePath)
	}

	manifest, err := exercise.LoadManifest(exercisePath)
	if err != nil {
		return "", nil, fmt.Errorf("load exercise manifest: %w", err)
	}
	return exercisePath, manifest, nil
}

func resolveBearerToken(token string) string {
	token = strings.TrimSpace(token)
	if token != "" {
		return token
	}
	return strings.TrimSpace(os.Getenv(BearerTokenEnvVar))
}

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
)

type PublishRequest struct {
	ExercisePath string
	OrgID        string
	ExerciseID   string
	Version      string
	Status       string
	BearerToken  string
}

func (r *Remote) PublishRemote(request PublishRequest) (string, error) {
	exercisePath, err := filepath.Abs(request.ExercisePath)
	if err != nil {
		return "", fmt.Errorf("resolve exercise path: %w", err)
	}

	// Package the exercise
	publicPath, privatePath, err := r.PackageFunc(exercisePath)
	if err != nil {
		return "", fmt.Errorf("package exercise: %w", err)
	}
	defer os.Remove(publicPath)
	defer os.Remove(privatePath)

	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)

	// Add fields
	if err := writer.WriteField("org_id", request.OrgID); err != nil {
		return "", fmt.Errorf("write org_id: %w", err)
	}
	if request.ExerciseID != "" {
		if err := writer.WriteField("exercise_id", request.ExerciseID); err != nil {
			return "", fmt.Errorf("write exercise_id: %w", err)
		}
	}
	if request.Version != "" {
		if err := writer.WriteField("version", request.Version); err != nil {
			return "", fmt.Errorf("write version: %w", err)
		}
	}
	if request.Status != "" {
		if err := writer.WriteField("status", request.Status); err != nil {
			return "", fmt.Errorf("write status: %w", err)
		}
	}

	// Add public_artifact file
	publicFilePart, err := writer.CreateFormFile("public_artifact", filepath.Base(publicPath))
	if err != nil {
		return "", fmt.Errorf("create public_artifact form file: %w", err)
	}
	publicFile, err := os.Open(publicPath)
	if err != nil {
		return "", fmt.Errorf("open public package: %w", err)
	}
	if _, err := io.Copy(publicFilePart, publicFile); err != nil {
		publicFile.Close()
		return "", fmt.Errorf("copy public package: %w", err)
	}
	publicFile.Close()

	// Add private_artifact file
	privateFilePart, err := writer.CreateFormFile("private_artifact", filepath.Base(privatePath))
	if err != nil {
		return "", fmt.Errorf("create private_artifact form file: %w", err)
	}
	privateFile, err := os.Open(privatePath)
	if err != nil {
		return "", fmt.Errorf("open private package: %w", err)
	}
	if _, err := io.Copy(privateFilePart, privateFile); err != nil {
		privateFile.Close()
		return "", fmt.Errorf("copy private package: %w", err)
	}
	privateFile.Close()

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("finalize publish form: %w", err)
	}

	publishURL, err := r.publishURL()
	if err != nil {
		return "", err
	}

	httpRequest, err := http.NewRequest(http.MethodPost, publishURL, bodyBuffer)
	if err != nil {
		return "", fmt.Errorf("create publish request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())

	bearerToken := resolveBearerToken(request.BearerToken)
	if bearerToken != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	response, err := r.httpClient().Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("publish package request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8192))
	if err != nil {
		return "", fmt.Errorf("read publish response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(responseBody))
		if message == "" {
			message = response.Status
		}
		return "", fmt.Errorf("publish exercise: %s", message)
	}

	return strings.TrimSpace(string(responseBody)), nil
}

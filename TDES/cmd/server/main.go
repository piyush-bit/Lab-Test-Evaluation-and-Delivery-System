package main

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"TDES/internals/registry"
	evaluatorcore "TDES/internals/evaluator-core"
)

type Config struct {
	Port         string
	DataRoot     string
	DBPath       string
	ArtifactRoot string
}

func loadConfig() Config {
	dataRoot := getEnv("REGISTRY_DATA_ROOT", "./data")
	return Config{
		Port:         getEnv("PORT", "8080"),
		DataRoot:     dataRoot,
		DBPath:       getEnv("REGISTRY_DB_PATH", filepath.Join(dataRoot, "registry.db")),
		ArtifactRoot: getEnv("REGISTRY_ARTIFACT_ROOT", filepath.Join(dataRoot, "objects")),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	cfg := loadConfig()

	log.Printf("Starting Registry Server...")
	log.Printf("  Data Root:     %s", cfg.DataRoot)
	log.Printf("  Database:      %s", cfg.DBPath)
	log.Printf("  Artifact Root: %s", cfg.ArtifactRoot)
	log.Printf("  Port:          %s", cfg.Port)

	// Ensure directories exist
	if err := os.MkdirAll(cfg.DataRoot, 0755); err != nil {
		log.Fatalf("Fatal: failed to create data root: %v", err)
	}

	repo, err := registry.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v", err)
	}
	defer repo.Close()

	store, err := registry.NewDiskArtifactStore(cfg.ArtifactRoot)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize artifact store: %v", err)
	}

	service, err := registry.NewService(repo, store)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize registry service: %v", err)
	}

	mux := http.NewServeMux()

	// 1. Healthcheck
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// 2. Publish
	mux.HandleFunc("POST /v1/exercises/publish", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if !isInstructorTokenValid(token) {
			respondError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		// Limit body size to 64MB for safety
		r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
			return
		}

		orgID := r.FormValue("org_id")
		exerciseID := r.FormValue("exercise_id")
		version := r.FormValue("version")
		status := r.FormValue("status")

		publicFile, _, err := r.FormFile("public_artifact")
		if err != nil {
			respondError(w, http.StatusBadRequest, "public_artifact is required")
			return
		}
		defer publicFile.Close()

		privateFile, _, err := r.FormFile("private_artifact")
		if err != nil {
			respondError(w, http.StatusBadRequest, "private_artifact is required")
			return
		}
		defer privateFile.Close()

		// Save uploads to temp files
		publicTemp, err := saveTempFile(publicFile)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save public artifact: "+err.Error())
			return
		}
		defer os.Remove(publicTemp)

		privateTemp, err := saveTempFile(privateFile)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save private artifact: "+err.Error())
			return
		}
		defer os.Remove(privateTemp)

		storedVersion, created, err := service.Publish(r.Context(), registry.PublishRequest{
			OrgID:               orgID,
			ExerciseID:          exerciseID,
			Version:             version,
			Status:              status,
			PublicArtifactPath:  publicTemp,
			PrivateArtifactPath: privateTemp,
		})
		if err != nil {
			code := http.StatusBadRequest
			if errors.Is(err, registry.ErrExerciseVersionConflict) {
				code = http.StatusConflict
			}
			respondError(w, code, err.Error())
			return
		}

		code := http.StatusCreated
		if !created {
			code = http.StatusOK
		}
		respondJSON(w, code, storedVersion)
	})

	// 3. Get Exercise Version
	mux.HandleFunc("GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.PathValue("orgID")
		exerciseID := r.PathValue("exerciseID")
		version := r.PathValue("version")

		ev, err := service.GetExerciseVersion(r.Context(), orgID, exerciseID, version)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "exercise version not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, ev)
	})

	// 4. Download Artifact by SHA
	mux.HandleFunc("GET /v1/artifacts/{sha256}", func(w http.ResponseWriter, r *http.Request) {
		sha := r.PathValue("sha256")
		handle, err := service.OpenArtifact(r.Context(), sha)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "artifact not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer handle.File.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("ETag", handle.Artifact.SHA256)
		w.Header().Set("X-Artifact-SHA256", handle.Artifact.SHA256)
		w.Header().Set("X-Object-Key", handle.Artifact.ObjectKey)
		w.Header().Set("Content-Length", strconv.FormatInt(handle.Artifact.SizeBytes, 10))
		w.WriteHeader(http.StatusOK)

		_, _ = io.Copy(w, handle.File)
	})

	// 5. Download Artifact by org/exercise/version
	mux.HandleFunc("GET /v1/exercises/{orgID}/{exerciseID}/versions/{version}/download", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.PathValue("orgID")
		exerciseID := r.PathValue("exerciseID")
		version := r.PathValue("version")
		downloadType := strings.ToLower(r.URL.Query().Get("type"))
		if downloadType != "private" {
			downloadType = "public"
		}

		ev, err := service.GetExerciseVersion(r.Context(), orgID, exerciseID, version)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "exercise version not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sha := ev.PublicArtifactSHA
		if downloadType == "private" {
			sha = ev.PrivateArtifactSHA
		}

		handle, err := service.OpenArtifact(r.Context(), sha)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "artifact file not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer handle.File.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("ETag", handle.Artifact.SHA256)
		w.Header().Set("X-Artifact-SHA256", handle.Artifact.SHA256)
		w.Header().Set("X-Object-Key", handle.Artifact.ObjectKey)
		w.Header().Set("Content-Length", strconv.FormatInt(handle.Artifact.SizeBytes, 10))
		w.WriteHeader(http.StatusOK)

		_, _ = io.Copy(w, handle.File)
	})

	// 6. List Exercises
	mux.HandleFunc("GET /v1/exercises", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.URL.Query().Get("org_id")
		status := r.URL.Query().Get("status")

		list, err := service.ListExercises(r.Context(), orgID, status)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, list)
	})

	// 7. List Versions of an Exercise
	mux.HandleFunc("GET /v1/exercises/{orgID}/{exerciseID}/versions", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.PathValue("orgID")
		exerciseID := r.PathValue("exerciseID")

		list, err := service.ListExerciseVersions(r.Context(), orgID, exerciseID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, list)
	})

	// 8. Update Status
	mux.HandleFunc("POST /v1/exercises/{orgID}/{exerciseID}/versions/{version}/status", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.PathValue("orgID")
		exerciseID := r.PathValue("exerciseID")
		version := r.PathValue("version")
		status := r.FormValue("status")
		if status == "" {
			status = r.URL.Query().Get("status")
		}
		if status == "" {
			// Try decoding from json body
			var body struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				status = body.Status
			}
		}

		if strings.TrimSpace(status) == "" {
			respondError(w, http.StatusBadRequest, "status is required")
			return
		}

		err := service.UpdateExerciseStatus(r.Context(), orgID, exerciseID, version, status)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "exercise version not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// 9. Delete Exercise Version
	mux.HandleFunc("DELETE /v1/exercises/{orgID}/{exerciseID}/versions/{version}", func(w http.ResponseWriter, r *http.Request) {
		orgID := r.PathValue("orgID")
		exerciseID := r.PathValue("exerciseID")
		version := r.PathValue("version")

		err := service.DeleteExerciseVersion(r.Context(), orgID, exerciseID, version)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusNotFound, "exercise version not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// 10. Remote Submission Evaluator
	mux.HandleFunc("POST /v1/submissions", handleSubmissions(service, nil))

	// 11. Get/Export Submissions
	mux.HandleFunc("GET /v1/submissions", handleGetSubmissions(service))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on port %s...", cfg.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server listen error: %v", err)
	}
	log.Println("Server stopped.")
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func saveTempFile(src io.Reader) (string, error) {
	tmpFile, err := os.CreateTemp("", "upload-*.tar")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

type serviceArtifactProvider struct {
	service *registry.Service
}

func (p *serviceArtifactProvider) OpenPrivateArtifact(ctx context.Context, orgID, labID, version string) (io.ReadCloser, error) {
	ev, err := p.service.GetExerciseVersion(ctx, orgID, labID, version)
	if err != nil {
		return nil, err
	}
	handle, err := p.service.OpenArtifact(ctx, ev.PrivateArtifactSHA)
	if err != nil {
		return nil, err
	}
	return handle.File, nil
}

func handleSubmissions(service *registry.Service, runtime evaluatorcore.Runtime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Limit body size to 64MB for safety
		r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
			return
		}

		authHeader := r.Header.Get("Authorization")
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		expectedToken := os.Getenv("EUC2_REMOTE_BEARER_TOKEN")
		if expectedToken != "" && token != expectedToken {
			respondError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		submissionFile, _, err := r.FormFile("submission_package")
		if err != nil {
			respondError(w, http.StatusBadRequest, "submission_package is required")
			return
		}
		defer submissionFile.Close()

		tempPath, err := saveTempFile(submissionFile)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save submission package: "+err.Error())
			return
		}
		defer os.Remove(tempPath)

		// PIN & TOFU Verification
		studentID, orgID, err := readSubmissionIdentity(tempPath)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid submission package: "+err.Error())
			return
		}

		pin := strings.TrimSpace(r.FormValue("pin"))
		newPin := strings.TrimSpace(r.FormValue("new_pin"))

		cred, err := service.GetStudentCredential(r.Context(), orgID, studentID)
		if err != nil {
			if errors.Is(err, registry.ErrNotFound) {
				respondError(w, http.StatusForbidden, fmt.Sprintf("student %s not registered in roster for org %s", studentID, orgID))
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to check student roster: "+err.Error())
			return
		}

		if cred.PinHash == "" {
			// TOFU Activation phase
			if len(pin) < 4 {
				respondError(w, http.StatusBadRequest, "PIN must be at least 4 characters long")
				return
			}
			cred.PinHash = hashPin(studentID, pin)
			if err := service.SaveStudentCredential(r.Context(), cred); err != nil {
				respondError(w, http.StatusInternalServerError, "failed to register student PIN: "+err.Error())
				return
			}
			log.Printf("Student %s/%s registered PIN successfully", orgID, studentID)
		} else {
			// Validation phase
			expectedHash := hashPin(studentID, pin)
			if cred.PinHash != expectedHash {
				respondError(w, http.StatusUnauthorized, "invalid student PIN")
				return
			}

			// PIN Update phase
			if newPin != "" {
				if len(newPin) < 4 {
					respondError(w, http.StatusBadRequest, "new PIN must be at least 4 characters long")
					return
				}
				cred.PinHash = hashPin(studentID, newPin)
				if err := service.SaveStudentCredential(r.Context(), cred); err != nil {
					respondError(w, http.StatusInternalServerError, "failed to update student PIN: "+err.Error())
					return
				}
				log.Printf("Student %s/%s updated PIN successfully", orgID, studentID)
			}
		}

		provider := &serviceArtifactProvider{service: service}
		evaluator, err := evaluatorcore.NewEvaluator(provider, runtime)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create evaluator: "+err.Error())
			return
		}

		log.Printf("Received remote submission upload for evaluation")
		result, err := evaluator.EvaluateSubmission(r.Context(), evaluatorcore.EvaluationRequest{
			SubmissionArchivePath: tempPath,
			DockerBinary:          os.Getenv("DOCKER_BINARY"),
		})
		if err != nil {
			log.Printf("Evaluation error: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to evaluate submission: "+err.Error())
			return
		}

		log.Printf("Evaluation completed: OrgID=%s, StudentID=%s, LabID=%s, Version=%s, Status=%s, Score=%d/%d",
			result.OrgID, result.StudentID, result.LabID, result.Version, result.Status, result.EarnedPoints, result.MaxPoints)
		for _, testResult := range result.Results {
			log.Printf("  - Command: %q -> Status: %s, Points: %d/%d",
				testResult.Command, testResult.Status, testResult.PointsEarned, testResult.PointsPossible)
		}

		resultsBytes, _ := json.Marshal(result.Results)
		evalRecord := registry.SubmissionEvaluation{
			OrgID:        result.OrgID,
			StudentID:    result.StudentID,
			LabID:        result.LabID,
			Version:      result.Version,
			Status:       result.Status,
			EarnedPoints: result.EarnedPoints,
			MaxPoints:    result.MaxPoints,
			ResultsJSON:  string(resultsBytes),
		}

		if err := service.SaveEvaluation(r.Context(), evalRecord); err != nil {
			log.Printf("Failed to persist evaluation record: %v", err)
			respondError(w, http.StatusInternalServerError, "failed to save evaluation record: "+err.Error())
			return
		}

		respondJSON(w, http.StatusOK, result)
	}
}

func handleGetSubmissions(service *registry.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if !isInstructorTokenValid(token) {
			respondError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		orgID := r.URL.Query().Get("org_id")
		labID := r.URL.Query().Get("lab_id")
		format := strings.ToLower(r.URL.Query().Get("format"))

		submissions, err := service.ListSubmissions(r.Context(), orgID, labID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to query submissions: "+err.Error())
			return
		}

		if format == "csv" {
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", `attachment; filename="submissions.csv"`)
			w.WriteHeader(http.StatusOK)

			writer := csv.NewWriter(w)
			defer writer.Flush()

			header := []string{"id", "org_id", "student_id", "lab_id", "version", "status", "earned_points", "max_points", "created_at"}
			if err := writer.Write(header); err != nil {
				log.Printf("Error writing CSV header: %v", err)
				return
			}

			for _, s := range submissions {
				row := []string{
					s.ID,
					s.OrgID,
					s.StudentID,
					s.LabID,
					s.Version,
					s.Status,
					strconv.Itoa(s.EarnedPoints),
					strconv.Itoa(s.MaxPoints),
					s.CreatedAt.Format(time.RFC3339),
				}
				if err := writer.Write(row); err != nil {
					log.Printf("Error writing CSV row: %v", err)
					return
				}
			}
			return
		}

		respondJSON(w, http.StatusOK, submissions)
	}
}

func isInstructorTokenValid(token string) bool {
	instructorTokensStr := os.Getenv("EUC2_INSTRUCTOR_TOKENS")
	if instructorTokensStr == "" {
		// Fallback to EUC2_REMOTE_BEARER_TOKEN if instructor tokens are not configured
		fallback := os.Getenv("EUC2_REMOTE_BEARER_TOKEN")
		return fallback == "" || token == fallback
	}

	tokens := strings.Split(instructorTokensStr, ",")
	for _, t := range tokens {
		if token == strings.TrimSpace(t) {
			return true
		}
	}
	return false
}

func handleAdminOnboard(service *registry.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if !isInstructorTokenValid(token) {
			respondError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		var reader io.Reader
		// Try parsing as multipart form first
		if err := r.ParseMultipartForm(32 << 20); err == nil {
			file, _, err := r.FormFile("roster_csv")
			if err != nil {
				respondError(w, http.StatusBadRequest, "roster_csv file is required in multipart form")
				return
			}
			defer file.Close()
			reader = file
		} else {
			// Fallback to reading raw body
			reader = r.Body
		}

		csvReader := csv.NewReader(reader)
		records, err := csvReader.ReadAll()
		if err != nil {
			respondError(w, http.StatusBadRequest, "failed to parse CSV: "+err.Error())
			return
		}

		if len(records) < 2 {
			respondError(w, http.StatusBadRequest, "CSV roster is empty or missing headers")
			return
		}

		header := records[0]
		studentIDCol := -1
		orgIDCol := -1

		for idx, h := range header {
			hClean := strings.TrimSpace(strings.ToLower(h))
			if hClean == "student_id" || hClean == "studentid" || hClean == "student" {
				studentIDCol = idx
			}
			if hClean == "org_id" || hClean == "orgid" || hClean == "org" || hClean == "organization" {
				orgIDCol = idx
			}
		}

		if studentIDCol == -1 {
			respondError(w, http.StatusBadRequest, "missing student_id column in CSV header")
			return
		}

		onboardedCount := 0
		for i := 1; i < len(records); i++ {
			row := records[i]
			if len(row) <= studentIDCol {
				continue
			}
			studentID := strings.TrimSpace(row[studentIDCol])
			if studentID == "" {
				continue
			}

			orgID := "default"
			if orgIDCol != -1 && len(row) > orgIDCol {
				cleanedOrg := strings.TrimSpace(row[orgIDCol])
				if cleanedOrg != "" {
					orgID = cleanedOrg
				}
			}

			existing, err := service.GetStudentCredential(r.Context(), orgID, studentID)
			if err != nil {
				if errors.Is(err, registry.ErrNotFound) {
					err = service.SaveStudentCredential(r.Context(), registry.StudentCredential{
						OrgID:     orgID,
						StudentID: studentID,
					})
					if err != nil {
						log.Printf("Failed to onboard student %s/%s: %v", orgID, studentID, err)
						continue
					}
					onboardedCount++
				} else {
					log.Printf("Error checking student credential %s/%s: %v", orgID, studentID, err)
				}
			} else {
				log.Printf("Student %s/%s already rostered, skipping", orgID, studentID)
				_ = existing
			}
		}

		respondJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"onboarded": onboardedCount,
		})
	}
}

func readSubmissionIdentity(path string) (studentID, orgID string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	tr := tar.NewReader(file)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", err
		}
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == "submission-manifest.json" {
			var manifest struct {
				OrgID     string `json:"org_id"`
				StudentID string `json:"student_id"`
			}
			if err := json.NewDecoder(tr).Decode(&manifest); err != nil {
				return "", "", err
			}
			return manifest.StudentID, manifest.OrgID, nil
		}
	}
	return "", "", errors.New("submission-manifest.json not found in archive")
}

func hashPin(studentID, pin string) string {
	salt := os.Getenv("EUC2_PIN_SALT")
	if salt == "" {
		salt = "default-tdes-salt-value-for-security"
	}
	hash := sha256.Sum256([]byte(studentID + ":" + pin + ":" + salt))
	return hex.EncodeToString(hash[:])
}


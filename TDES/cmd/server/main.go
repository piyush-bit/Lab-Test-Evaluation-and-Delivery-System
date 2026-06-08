package main

import (
	"context"
	"encoding/json"
	"errors"
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

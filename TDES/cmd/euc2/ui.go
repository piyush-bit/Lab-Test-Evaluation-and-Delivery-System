package main

import (
	initMod "TDES/internals/init"
	exercisestore "TDES/internals/exercise_store"
	drive "TDES/internals/drive"
	exercise "TDES/internals/exercise"
	evaluatorcore "TDES/internals/evaluator-core"
	runtests "TDES/internals/run"
	"crypto/ecdh"
	cryptoRand "crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

//go:embed all:ui_dist
var uiFS embed.FS

var (
	uiPort string
	uiHost string
	noOpen bool
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the local TDES Web UI dashboard",
	Long:  `Start a local web server serving the embedded TDES Web Console.`,
	Run: func(cmd *cobra.Command, args []string) {
		mux := http.NewServeMux()

		// 1. Serve embedded static files with SPA fallback to index.html for unknown frontend routes
		subFS, err := fs.Sub(uiFS, "ui_dist")
		if err != nil {
			log.Fatalf("Fatal: failed to open embedded UI filesystem: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))

		// Catch-all handler for static files & SPA fallback
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Don't handle API routes here
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Clean path
			upath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
			if upath == "" {
				fileServer.ServeHTTP(w, r)
				return
			}

			// Try to open file in subFS
			f, err := subFS.Open(upath)
			if err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// If file does not exist, serve index.html for SPA routing
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
		})

		// 2. API: Status check with Docker status check
		mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
			cwd, _ := os.Getwd()
			dockerRunning := checkDockerStatus()
			
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":         "ok",
				"workspace":      cwd,
				"docker_running": dockerRunning,
			})
		})

		// 3. API: Directory Browser API
		mux.HandleFunc("GET /api/browse", func(w http.ResponseWriter, r *http.Request) {
			pathQuery := r.URL.Query().Get("path")
			
			if strings.HasPrefix(pathQuery, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					pathQuery = filepath.Join(home, strings.TrimPrefix(pathQuery, "~"))
				}
			}

			var targetPath string
			var err error
			if pathQuery == "" {
				targetPath, err = os.Getwd()
				if err != nil {
					targetPath, _ = os.UserHomeDir()
				}
			} else {
				targetPath, err = filepath.Abs(pathQuery)
				if err != nil {
					respondError(w, http.StatusBadRequest, "Invalid path: "+err.Error())
					return
				}
			}

			files, err := os.ReadDir(targetPath)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Could not read directory: "+err.Error())
				return
			}

			var dirs []string
			for _, file := range files {
				// Hide hidden directories starting with "."
				if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
					dirs = append(dirs, file.Name())
				}
			}

			parentPath := filepath.Dir(targetPath)
			if parentPath == targetPath {
				parentPath = ""
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"current_path": targetPath,
				"parent_path":  parentPath,
				"directories":  dirs,
			})
		})

		// 4. API: Validate Workspace
		mux.HandleFunc("GET /api/validate-workspace", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get("path")
			if strings.HasPrefix(path, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					path = filepath.Join(home, strings.TrimPrefix(path, "~"))
				}
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{
					"valid": false,
					"error": "Invalid path format",
				})
				return
			}

			info, err := os.Stat(absPath)
			if err != nil || !info.IsDir() {
				respondJSON(w, http.StatusOK, map[string]any{
					"valid": false,
					"error": "Directory does not exist",
				})
				return
			}

			// Validate workspace using ExerciseManifest struct matching
			manifest, err := exercise.LoadManifest(absPath)
			if err != nil || manifest == nil || strings.TrimSpace(manifest.LabID) == "" {
				respondJSON(w, http.StatusOK, map[string]any{
					"valid": false,
					"error": "directory does not contain a valid exercise workspace manifest",
				})
				return
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"valid":    true,
				"path":     absPath,
				"manifest": manifest,
			})
		})

		// 4b. API: List Cached Exercises
		mux.HandleFunc("GET /api/exercises", func(w http.ResponseWriter, r *http.Request) {
			index, err := exercisestore.LoadIndex(exercisestore.GetPublicCacheDir())
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, index)
		})

		// 4c. API: Validate Drive Path
		mux.HandleFunc("GET /api/validate-drive", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get("path")
			if path == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{"valid": false, "error": err.Error()})
				return
			}
			manifest, err := drive.ReadManifest(absPath)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{"valid": false, "error": err.Error()})
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"valid":    true,
				"path":     absPath,
				"manifest": manifest,
			})
		})

		// 4d. API: List Exercises on a Drive
		mux.HandleFunc("GET /api/drive-exercises", func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Query().Get("path")
			if path == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			index, err := exercisestore.LoadIndex(filepath.Join(absPath, "exercise"))
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, index)
		})

		// 4e. API: List Exercises on a Remote Registry
		mux.HandleFunc("GET /api/remote-exercises", func(w http.ResponseWriter, r *http.Request) {
			remoteURL := r.URL.Query().Get("remote_url")
			orgID := r.URL.Query().Get("org_id")
			search := r.URL.Query().Get("q")
			if remoteURL == "" {
				respondError(w, http.StatusBadRequest, "remote_url parameter is required")
				return
			}
			if orgID == "" {
				orgID = "default"
			}

			parsedURL, err := url.Parse(remoteURL)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid remote URL: "+err.Error())
				return
			}
			parsedURL.Path = path.Join(parsedURL.Path, "v1", "exercises")

			q := parsedURL.Query()
			q.Set("org_id", orgID)
			q.Set("status", "published")
			if search != "" {
				q.Set("q", search)
			}
			parsedURL.RawQuery = q.Encode()

			resp, err := http.Get(parsedURL.String())
			if err != nil {
				respondError(w, http.StatusBadGateway, "Failed to connect to remote registry: "+err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				respondError(w, resp.StatusCode, fmt.Sprintf("Remote registry returned error (%d): %s", resp.StatusCode, string(bodyBytes)))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.Copy(w, resp.Body)
		})

		// 4f. API: Check Remote Server Health (pings <remoteURL>/healthz)
		mux.HandleFunc("GET /api/remote/health", func(w http.ResponseWriter, r *http.Request) {
			rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
			if rawURL == "" {
				respondJSON(w, http.StatusOK, map[string]any{"online": false, "error": "url parameter is required"})
				return
			}

			if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
				rawURL = "http://" + rawURL
			}

			parsedURL, err := url.Parse(rawURL)
			if err != nil || parsedURL.Host == "" {
				respondJSON(w, http.StatusOK, map[string]any{
					"online": false,
					"url":    rawURL,
					"error":  "invalid URL format",
				})
				return
			}

			healthURL := fmt.Sprintf("%s://%s/healthz", parsedURL.Scheme, parsedURL.Host)
			client := &http.Client{Timeout: 3 * time.Second}
			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, healthURL, nil)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{
					"online": false,
					"url":    rawURL,
					"error":  err.Error(),
				})
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{
					"online": false,
					"url":    rawURL,
					"error":  "unreachable",
				})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				respondJSON(w, http.StatusOK, map[string]any{
					"online": true,
					"url":    rawURL,
				})
			} else {
				respondJSON(w, http.StatusOK, map[string]any{
					"online": false,
					"url":    rawURL,
					"error":  fmt.Sprintf("HTTP %d", resp.StatusCode),
				})
			}
		})

		// 4f. API: Admin - Prepare Drive
		mux.HandleFunc("POST /api/drive/prepare", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath string `json:"drive_path"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" {
				respondError(w, http.StatusBadRequest, "drive_path is required")
				return
			}
			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			if err := drive.PrepareDrive(drivePath); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to prepare drive: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message": "Drive prepared successfully",
				"path":    drivePath,
			})
		})

		// 4g. API: Admin - Prepare Drive Submission
		mux.HandleFunc("POST /api/drive/prepare-submission", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath          string `json:"drive_path"`
				RecipientPublicKey string `json:"recipient_public_key"`
				SubmissionID       string `json:"submission_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" {
				respondError(w, http.StatusBadRequest, "drive_path is required")
				return
			}
			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			manifest, generatedPrivateKey, err := drive.PrepareDriveForSubmissionWithID(drivePath, req.RecipientPublicKey, req.SubmissionID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to prepare drive submission module: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message":              "Drive submission module prepared successfully",
				"path":                 drivePath,
				"submission_id":        manifest.SubmissionID,
				"recipient_public_key": manifest.RecipientPublicKeyB64,
				"private_key":          generatedPrivateKey,
			})
		})

		// 4h. API: Admin - Add Exercise to Drive
		mux.HandleFunc("POST /api/drive/add-exercise", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath  string `json:"drive_path"`
				ExerciseID string `json:"exercise_id"`
				Version    string `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" || req.ExerciseID == "" || req.Version == "" {
				respondError(w, http.StatusBadRequest, "drive_path, exercise_id, and version are required")
				return
			}
			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			d := &drive.Drive{Path: drivePath}
			if err := d.AddExerciseFromID(req.ExerciseID, req.Version); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to add exercise to drive: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message":     fmt.Sprintf("Exercise %s@%s added to drive successfully", req.ExerciseID, req.Version),
				"exercise_id": req.ExerciseID,
				"version":     req.Version,
				"path":        drivePath,
			})
		})

		// 4i. API: Admin - Delete Exercise from Drive
		mux.HandleFunc("POST /api/drive/delete-exercise", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath  string `json:"drive_path"`
				ExerciseID string `json:"exercise_id"`
				Version    string `json:"version"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" || req.ExerciseID == "" {
				respondError(w, http.StatusBadRequest, "drive_path and exercise_id are required")
				return
			}
			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			storeRoot := filepath.Join(drivePath, "exercise")
			if err := exercisestore.RemoveExercise(storeRoot, req.ExerciseID, req.Version); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to delete exercise from drive: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message": fmt.Sprintf("Exercise %s deleted successfully from drive", req.ExerciseID),
			})
		})

		// 4j. API: Admin - Get Drive Submissions
		mux.HandleFunc("GET /api/drive/submissions", func(w http.ResponseWriter, r *http.Request) {
			drivePath := r.URL.Query().Get("path")
			if drivePath == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}
			absPath, err := filepath.Abs(drivePath)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			manifestPath := filepath.Join(absPath, "submissions", "manifest.json")
			manifestData, err := os.ReadFile(manifestPath)
			if err != nil {
				respondJSON(w, http.StatusOK, map[string]any{
					"prepared": false,
					"path":     absPath,
				})
				return
			}

			var subManifest drive.SubmissionManifest
			_ = json.Unmarshal(manifestData, &subManifest)

			keyRec, hasKey := drive.GetKeyRecord(subManifest.SubmissionID)

			submissionsDir := filepath.Join(absPath, "submissions")
			type submissionFile struct {
				Filename string    `json:"filename"`
				Size     int64     `json:"size"`
				ModTime  time.Time `json:"mod_time"`
				RelPath  string    `json:"rel_path"`
			}
			var files []submissionFile

			_ = filepath.Walk(submissionsDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if strings.HasSuffix(strings.ToLower(info.Name()), ".json") && info.Name() != "manifest.json" {
					rel, _ := filepath.Rel(submissionsDir, path)
					files = append(files, submissionFile{
						Filename: info.Name(),
						Size:     info.Size(),
						ModTime:  info.ModTime(),
						RelPath:  rel,
					})
				}
				return nil
			})

			respondJSON(w, http.StatusOK, map[string]any{
				"prepared":        true,
				"path":            absPath,
				"submission_id":   subManifest.SubmissionID,
				"public_key":      subManifest.RecipientPublicKeyB64,
				"has_private_key": hasKey,
				"private_key":     keyRec.PrivateKey,
				"manifest":        subManifest,
				"submissions":     files,
			})
		})

		// 4k. API: Admin - Save Key Record
		mux.HandleFunc("POST /api/drive/save-key", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				SubmissionID string `json:"submission_id"`
				PublicKey    string `json:"public_key"`
				PrivateKey   string `json:"private_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.SubmissionID == "" || req.PrivateKey == "" {
				respondError(w, http.StatusBadRequest, "submission_id and private_key are required")
				return
			}
			err := drive.SaveKeyRecord(drive.KeyRecord{
				SubmissionID: req.SubmissionID,
				PublicKey:    req.PublicKey,
				PrivateKey:   req.PrivateKey,
				CreatedAt:    time.Now().Format(time.RFC3339),
			})
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to save private key: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message":       "Private key saved successfully for submission ID " + req.SubmissionID,
				"submission_id": req.SubmissionID,
			})
		})

		// 4l. API: Admin - Generate Keypair
		mux.HandleFunc("POST /api/drive/generate-keypair", func(w http.ResponseWriter, r *http.Request) {
			privKey, err := ecdh.X25519().GenerateKey(cryptoRand.Reader)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to generate keypair: "+err.Error())
				return
			}
			pubKeyB64 := base64.StdEncoding.EncodeToString(privKey.PublicKey().Bytes())
			privKeyB64 := base64.StdEncoding.EncodeToString(privKey.Bytes())

			respondJSON(w, http.StatusOK, map[string]string{
				"public_key":  pubKeyB64,
				"private_key": privKeyB64,
			})
		})

		// 4l2. API: Admin - Clear Submissions Directory on Drive
		mux.HandleFunc("POST /api/drive/clear-submissions", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath string `json:"drive_path"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" {
				respondError(w, http.StatusBadRequest, "drive_path is required")
				return
			}

			if err := drive.ClearSubmissions(req.DrivePath); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to clear submissions: "+err.Error())
				return
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"message": "Submissions cleared successfully",
			})
		})

		// 4m. API: Admin - Batch Evaluate Submissions
		mux.HandleFunc("POST /api/drive/evaluate-batch", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath           string `json:"drive_path"`
				RecipientPrivateKey string `json:"recipient_private_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" {
				respondError(w, http.StatusBadRequest, "drive_path is required")
				return
			}

			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			if abs, err := filepath.Abs(drivePath); err == nil {
				drivePath = abs
			}

			privKeyStr := strings.TrimSpace(req.RecipientPrivateKey)
			if privKeyStr == "" {
				manifestPath := filepath.Join(drivePath, "submissions", "manifest.json")
				manifestData, err := os.ReadFile(manifestPath)
				if err == nil {
					var subManifest drive.SubmissionManifest
					if err := json.Unmarshal(manifestData, &subManifest); err == nil && subManifest.SubmissionID != "" {
						if keyRec, ok := drive.GetKeyRecord(subManifest.SubmissionID); ok && keyRec.PrivateKey != "" {
							privKeyStr = keyRec.PrivateKey
						}
					}
				}
			}

			if privKeyStr == "" {
				respondError(w, http.StatusBadRequest, "recipient_private_key is required (or private key must be saved in key store)")
				return
			}

			rawKey, err := base64.StdEncoding.DecodeString(privKeyStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid recipient private key base64: "+err.Error())
				return
			}
			privateKey, err := ecdh.X25519().NewPrivateKey(rawKey)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid X25519 private key: "+err.Error())
				return
			}

			decrypted, err := drive.LoadAndDecryptSubmissions(drivePath, privateKey)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to load submissions: "+err.Error())
				return
			}

			var records []BatchResultRecord
			for _, sub := range decrypted {
				record := BatchResultRecord{
					EnvelopePath: sub.EnvelopePath,
					Status:       "error",
				}

				if sub.Error != nil {
					record.Error = fmt.Sprintf("decryption failed: %v", sub.Error)
					records = append(records, record)
					continue
				}

				tempFile, err := os.CreateTemp("", "batch-eval-*.tar")
				if err != nil {
					record.Error = fmt.Sprintf("create temp file: %v", err)
					records = append(records, record)
					continue
				}
				tempPath := tempFile.Name()
				_, _ = tempFile.Write(sub.PlaintextTar)
				tempFile.Close()

				resultJSON, err := evaluateSubmissionFile(r.Context(), tempPath, evaluationOptions{
					PrivateStore: filepath.Join(drivePath, "exercise"),
				})
				os.Remove(tempPath)

				if err != nil {
					record.Error = fmt.Sprintf("evaluation failed: %v", err)
					records = append(records, record)
					continue
				}

				var evalResult evaluatorcore.EvaluationResult
				if err := json.Unmarshal(resultJSON, &evalResult); err != nil {
					record.Error = fmt.Sprintf("decode result JSON: %v", err)
					records = append(records, record)
					continue
				}

				record.StudentID = evalResult.StudentID
				record.LabID = evalResult.LabID
				record.Version = evalResult.Version
				record.Status = evalResult.Status
				record.EarnedPoints = evalResult.EarnedPoints
				record.MaxPoints = evalResult.MaxPoints
				record.Results = evalResult.Results
				records = append(records, record)
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"records": records,
				"count":   len(records),
			})
		})

		// 4n. API: Admin - Evaluate Single Submission
		mux.HandleFunc("POST /api/drive/evaluate-single", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				DrivePath           string `json:"drive_path"`
				Filename            string `json:"filename"`
				RecipientPrivateKey string `json:"recipient_private_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" || req.Filename == "" {
				respondError(w, http.StatusBadRequest, "drive_path and filename are required")
				return
			}

			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			if abs, err := filepath.Abs(drivePath); err == nil {
				drivePath = abs
			}

			privKeyStr := strings.TrimSpace(req.RecipientPrivateKey)
			if privKeyStr == "" {
				manifestPath := filepath.Join(drivePath, "submissions", "manifest.json")
				manifestData, err := os.ReadFile(manifestPath)
				if err == nil {
					var subManifest drive.SubmissionManifest
					if err := json.Unmarshal(manifestData, &subManifest); err == nil && subManifest.SubmissionID != "" {
						if keyRec, ok := drive.GetKeyRecord(subManifest.SubmissionID); ok && keyRec.PrivateKey != "" {
							privKeyStr = keyRec.PrivateKey
						}
					}
				}
			}

			if privKeyStr == "" {
				respondError(w, http.StatusBadRequest, "recipient_private_key is required (or private key must be saved in key store)")
				return
			}

			var envelopePath string
			var envelopeData []byte

			submissionsDir := filepath.Join(drivePath, "submissions")
			_ = filepath.Walk(submissionsDir, func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				relSub, _ := filepath.Rel(submissionsDir, p)
				relDrive, _ := filepath.Rel(drivePath, p)
				if info.Name() == req.Filename || relSub == req.Filename || relDrive == req.Filename || p == req.Filename {
					if data, err := os.ReadFile(p); err == nil {
						envelopePath = p
						envelopeData = data
						return io.EOF
					}
				}
				return nil
			})

			if envelopeData == nil {
				respondError(w, http.StatusNotFound, fmt.Sprintf("Submission file not found (%s)", req.Filename))
				return
			}

			var envelope drive.SubmissionEnvelope
			if err := json.Unmarshal(envelopeData, &envelope); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid submission JSON: "+err.Error())
				return
			}

			rawKey, err := base64.StdEncoding.DecodeString(privKeyStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid private key base64: "+err.Error())
				return
			}
			privateKey, err := ecdh.X25519().NewPrivateKey(rawKey)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid X25519 private key: "+err.Error())
				return
			}

			plaintext, err := drive.DecryptSubmissionArchive(envelope, privateKey)
			if err != nil {
				respondError(w, http.StatusBadRequest, "Decryption failed: "+err.Error())
				return
			}

			tempFile, err := os.CreateTemp("", "single-eval-*.tar")
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Temp file creation failed: "+err.Error())
				return
			}
			tempPath := tempFile.Name()
			_, _ = tempFile.Write(plaintext)
			tempFile.Close()
			defer os.Remove(tempPath)

			resultJSON, err := evaluateSubmissionFile(r.Context(), tempPath, evaluationOptions{
				PrivateStore: filepath.Join(drivePath, "exercise"),
			})
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Evaluation failed: "+err.Error())
				return
			}

			var evalResult evaluatorcore.EvaluationResult
			_ = json.Unmarshal(resultJSON, &evalResult)

			record := BatchResultRecord{
				EnvelopePath: envelopePath,
				StudentID:    evalResult.StudentID,
				LabID:        evalResult.LabID,
				Version:      evalResult.Version,
				Status:       evalResult.Status,
				EarnedPoints: evalResult.EarnedPoints,
				MaxPoints:    evalResult.MaxPoints,
				Results:      evalResult.Results,
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"record": record,
				"result": evalResult,
			})
		})

		// 5. API: Fetch Exercise (Required for creating workspace)
		mux.HandleFunc("POST /api/fetch", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				ExerciseID string `json:"exercise_id"`
				Version    string `json:"version"`
				RemoteURL  string `json:"remote_url"`
				OrgID      string `json:"org_id"`
				DrivePath  string `json:"drive_path"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request JSON")
				return
			}

			var source fetchSource
			if req.DrivePath != "" {
				source = fetchSource{name: "drive", path: req.DrivePath}
			} else if req.RemoteURL != "" {
				source = fetchSource{name: "remote", path: req.RemoteURL}
			} else {
				respondError(w, http.StatusBadRequest, "Either remote_url or drive_path is required")
				return
			}

			fetchOrgID = req.OrgID
			err := fetchExercise(source, req.ExerciseID, req.Version)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "Exercise fetched successfully"})
		})

		// 6. API: Init Workspace
		mux.HandleFunc("POST /api/init", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				ExerciseID string `json:"exercise_id"`
				Version    string `json:"version"`
				TargetDir  string `json:"target_dir"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request JSON")
				return
			}

			err := initMod.InitFromID(req.ExerciseID, req.Version, req.TargetDir)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]string{"message": "Workspace initialized successfully"})
		})

		// 7. API: Get Workspace File Tree
		mux.HandleFunc("GET /api/workspace/files", func(w http.ResponseWriter, r *http.Request) {
			dirPath := r.URL.Query().Get("path")
			if dirPath == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}

			tree, err := buildFileTree(dirPath)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			respondJSON(w, http.StatusOK, tree)
		})

		// 8. API: Get Workspace File Content
		mux.HandleFunc("GET /api/workspace/file", func(w http.ResponseWriter, r *http.Request) {
			filePath := r.URL.Query().Get("path")
			if filePath == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			respondJSON(w, http.StatusOK, map[string]string{"content": string(content)})
		})

		// 9. API: Save Workspace File Content
		mux.HandleFunc("POST /api/workspace/file", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request JSON")
				return
			}

			if req.Path == "" {
				respondError(w, http.StatusBadRequest, "path is required")
				return
			}

			err := os.WriteFile(req.Path, []byte(req.Content), 0644)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			respondJSON(w, http.StatusOK, map[string]string{"message": "File saved successfully"})
		})

		// 10. API: Run Workspace Tests
		mux.HandleFunc("POST /api/workspace/run", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Path    string `json:"path"`
				Mode    string `json:"mode"`    // "docker" or "local"
				Command string `json:"command"` // optional, to run a single test command
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request JSON")
				return
			}

			if req.Path == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}

			config := runtests.Config{
				ExercisePath: req.Path,
			}

			result, err := runtests.RunGradingTests(config, req.Mode == "local", req.Command)
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}

			respondJSON(w, http.StatusOK, result)
		})

		// 6. API: Get Saved Submission Credentials
		mux.HandleFunc("GET /api/workspace/submit-config", func(w http.ResponseWriter, r *http.Request) {
			config, _ := loadLocalConfig()
			respondJSON(w, http.StatusOK, map[string]string{
				"student_id": config.StudentID,
				"org_id":     config.OrgID,
				"pin":        config.Pin,
			})
		})

		// 7. API: Submit Workspace (Remote Server or Drive)
		mux.HandleFunc("POST /api/workspace/submit", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Path      string `json:"path"`
				Strategy  string `json:"strategy"`  // "remote" or "drive"
				Target    string `json:"target"`    // remote URL or drive path
				StudentID string `json:"student_id"`
				OrgID     string `json:"org_id"`
				Pin       string `json:"pin"`
				NewPin    string `json:"new_pin"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid submission request payload")
				return
			}

			if req.Path == "" {
				respondError(w, http.StatusBadRequest, "path parameter is required")
				return
			}
			if req.Strategy != "remote" && req.Strategy != "drive" {
				respondError(w, http.StatusBadRequest, "strategy must be 'remote' or 'drive'")
				return
			}
			if strings.TrimSpace(req.Target) == "" {
				respondError(w, http.StatusBadRequest, "target destination (server URL or drive path) is required")
				return
			}

			studentID := strings.TrimSpace(req.StudentID)
			if studentID == "" {
				respondError(w, http.StatusBadRequest, "Student ID is required")
				return
			}
			orgID := strings.TrimSpace(req.OrgID)
			if orgID == "" {
				orgID = "default"
			}
			pin := strings.TrimSpace(req.Pin)
			if req.Strategy == "remote" && pin == "" {
				respondError(w, http.StatusBadRequest, "PIN code is required for remote submission")
				return
			}

			strat := submitStrategy{
				name: req.Strategy,
				path: req.Target,
			}

			resultStr, err := submitExerciseWithPath(req.Path, strat, orgID, studentID, pin, req.NewPin)
			if err != nil {
				respondError(w, http.StatusBadRequest, err.Error())
				return
			}

			// Save updated config locally
			config, configPath := loadLocalConfig()
			if configPath != "" {
				config.StudentID = studentID
				config.OrgID = orgID
				if req.Strategy == "remote" {
					if req.NewPin != "" {
						config.Pin = req.NewPin
					} else if pin != "" {
						config.Pin = pin
					}
				}
				saveLocalConfig(configPath, config)
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"result":  resultStr,
			})
		})

		serverAddr := uiHost + ":" + uiPort
		log.Printf("Starting TDES Web Console on http://%s ...", serverAddr)

		if !noOpen {
			go openBrowser("http://" + serverAddr)
		}

		if err := http.ListenAndServe(serverAddr, mux); err != nil {
			log.Fatalf("Server stopped with error: %v", err)
		}
	},
}

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

func buildFileTree(dirPath string) ([]FileNode, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var nodes []FileNode
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		node := FileNode{
			Name:  name,
			Path:  fullPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			children, err := buildFileTree(fullPath)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
		nodes = append(nodes, node)
	}

	if nodes == nil {
		nodes = []FileNode{}
	}
	return nodes, nil
}


func checkDockerStatus() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().StringVarP(&uiPort, "port", "p", "8082", "Port to bind local web console to")
	uiCmd.Flags().StringVarP(&uiHost, "host", "o", "127.0.0.1", "Host bind address")
	uiCmd.Flags().BoolVar(&noOpen, "no-open", false, "Disable auto-opening the browser")
}

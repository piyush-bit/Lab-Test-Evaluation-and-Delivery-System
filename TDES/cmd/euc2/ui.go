package main

import (
	initMod "TDES/internals/init"
	exercisestore "TDES/internals/exercise_store"
	drive "TDES/internals/drive"
	exercise "TDES/internals/exercise"
	runtests "TDES/internals/run"
	"embed"
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
				DrivePath           string `json:"drive_path"`
				RecipientPublicKey string `json:"recipient_public_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "Invalid request payload")
				return
			}
			if req.DrivePath == "" {
				respondError(w, http.StatusBadRequest, "drive_path is required")
				return
			}
			if req.RecipientPublicKey == "" {
				respondError(w, http.StatusBadRequest, "recipient_public_key is required")
				return
			}
			drivePath := req.DrivePath
			if strings.HasPrefix(drivePath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					drivePath = filepath.Join(home, strings.TrimPrefix(drivePath, "~"))
				}
			}
			if err := drive.PrepareDriveForSubmission(drivePath, req.RecipientPublicKey); err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to prepare drive submission module: "+err.Error())
				return
			}
			respondJSON(w, http.StatusOK, map[string]any{
				"message": "Drive submission module prepared successfully",
				"path":    drivePath,
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

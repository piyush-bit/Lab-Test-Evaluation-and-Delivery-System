package main

import (
	initMod "TDES/internals/init"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
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

		// 1. Serve embedded static files
		subFS, err := fs.Sub(uiFS, "ui_dist")
		if err != nil {
			log.Fatalf("Fatal: failed to open embedded UI filesystem: %v", err)
		}
		fileServer := http.FileServer(http.FS(subFS))
		mux.Handle("/", fileServer)

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

			// Validate workspace: check for manifest.json
			manifestPath := filepath.Join(absPath, "manifest.json")
			_, mErr := os.Stat(manifestPath)
			isValid := (mErr == nil)

			var manifestData any
			if isValid {
				mFile, mReadErr := os.ReadFile(manifestPath)
				if mReadErr == nil {
					_ = json.Unmarshal(mFile, &manifestData)
				}
			}

			respondJSON(w, http.StatusOK, map[string]any{
				"valid":    isValid,
				"path":     absPath,
				"manifest": manifestData,
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

package drive

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var Modules = []string{"drive", "evaluator", "delivery"}

type DriveManifest struct {
	Owner         string          `json:"owner"`
	ActiveModules map[string]bool `json:"active_modules"`
	LastSyncTime  time.Time       `json:"last_sync_time"`
}

type Drive struct {
	Manifest *DriveManifest
	Path     string
}

func ReadManifest(path string) (*DriveManifest, error) {
	manifestPath, err := manifestFilePath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read drive manifest %q: %w", manifestPath, err)
	}

	var manifest DriveManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode drive manifest %q: %w", manifestPath, err)
	}

	normalizeActiveModules(&manifest)
	if err := validateManifest(manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (m *DriveManifest) WriteManifest(path string) error {
	if m == nil {
		return fmt.Errorf("drive manifest is required")
	}

	manifestPath, err := manifestFilePath(path)
	if err != nil {
		return err
	}

	manifest := *m
	normalizeActiveModules(&manifest)
	if err := validateManifest(manifest); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode drive manifest: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return fmt.Errorf("create drive manifest directory: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("write drive manifest %q: %w", manifestPath, err)
	}
	return nil
}

func loadDriveManifest(path string, create bool) (*DriveManifest, error) {
	manifest, err := ReadManifest(path)
	if err == nil {
		return manifest, nil
	}

	var pathErr *os.PathError
	if !create || !errors.As(err, &pathErr) || !errors.Is(pathErr.Err, os.ErrNotExist) {
		return nil, err
	}

	manifest = defaultDriveManifest()
	if err := manifest.WriteManifest(path); err != nil {
		return nil, err
	}
	return manifest, nil
}

func defaultDriveManifest() *DriveManifest {
	manifest := &DriveManifest{
		ActiveModules: make(map[string]bool, len(Modules)),
	}
	normalizeActiveModules(manifest)
	return manifest
}

func normalizeActiveModules(manifest *DriveManifest) {
	if manifest.ActiveModules == nil {
		manifest.ActiveModules = make(map[string]bool, len(Modules))
	}
	for _, module := range Modules {
		if _, exists := manifest.ActiveModules[module]; !exists {
			manifest.ActiveModules[module] = false
		}
	}
}

func validateManifest(manifest DriveManifest) error {
	if manifest.ActiveModules == nil {
		return fmt.Errorf("drive manifest is missing active_modules")
	}

	allowedModules := make(map[string]struct{}, len(Modules))
	for _, module := range Modules {
		allowedModules[module] = struct{}{}
	}

	for module := range manifest.ActiveModules {
		if _, ok := allowedModules[module]; !ok {
			return fmt.Errorf("drive manifest has unknown module %q", module)
		}
	}

	return nil
}

func manifestFilePath(path string) (string, error) {
	driveRoot, err := resolveDriveRoot(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(driveRoot, "manifest.json"), nil
}

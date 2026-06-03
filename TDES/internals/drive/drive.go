package drive

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	exercisestore "TDES/internals/exercise_store"
)

func PrepareDrive(drivePath string) error {
	_, err := resolveDrive(drivePath, true)
	return err
}

func ResolveDrivePath(drivePath string) (*DriveManifest, error) {
	drive, err := resolveDrive(drivePath, false)
	if err != nil {
		return nil, err
	}
	return drive.Manifest, nil
}

func resolveDrive(drivePath string, create bool) (*Drive, error) {
	resolvedPath, err := resolveDriveRoot(drivePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("access drive path %q: %w", resolvedPath, err)
		}
		if !create {
			return nil, fmt.Errorf("drive path %q does not exist", resolvedPath)
		}
		if err := os.MkdirAll(resolvedPath, 0755); err != nil {
			return nil, fmt.Errorf("create drive path %q: %w", resolvedPath, err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("drive path %q is not a directory", resolvedPath)
	}

	manifest, err := loadDriveManifest(resolvedPath, create)
	if err != nil {
		return nil, err
	}

	if create {
		if _, err := exercisestore.LoadIndex(driveExerciseStoreRoot(resolvedPath)); err != nil {
			return nil, fmt.Errorf("prepare drive exercise store: %w", err)
		}
	}

	return &Drive{
		Manifest: manifest,
		Path:     resolvedPath,
	}, nil
}

func resolveDriveRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("drive path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve drive path: %w", err)
	}
	return absPath, nil
}

package drive

import (
	"fmt"
	"path/filepath"
	exercisestore "euc2/internals/exercise_store"
)

func (d *Drive) AddExerciseFromFile(packagePath string) error {
	driveRoot, err := d.exerciseStoreRoot()
	if err != nil {
		return err
	}
	if err := exercisestore.SavePackage(driveRoot, packagePath); err != nil {
		return fmt.Errorf("save exercise to drive: %w", err)
	}
	return nil
}

func (d *Drive) AddExerciseFromID(exerciseID string, exerciseVersion string) error {
	packagePath, err := exercisestore.ResolveExercisePath(exercisestore.GetPublicCacheDir(), exerciseID, exerciseVersion)
	if err != nil {
		return fmt.Errorf("resolve exercise from local cache: %w", err)
	}
	return d.AddExerciseFromFile(packagePath)
}

func (d *Drive) GetExerciseFile(exerciseID string, exerciseVersion string) (string, error) {
	driveRoot, err := d.exerciseStoreRoot()
	if err != nil {
		return "", err
	}
	path, err := exercisestore.ResolveExercisePath(driveRoot, exerciseID, exerciseVersion)
	if err != nil {
		return "", fmt.Errorf("resolve exercise from drive: %w", err)
	}
	return path, nil
}

func (d *Drive) exerciseStoreRoot() (string, error) {
	if d == nil {
		return "", fmt.Errorf("drive is required")
	}

	resolvedDrive, err := resolveDrive(d.Path, false)
	if err != nil {
		return "", err
	}

	d.Path = resolvedDrive.Path
	d.Manifest = resolvedDrive.Manifest
	return driveExerciseStoreRoot(d.Path), nil
}

func driveExerciseStoreRoot(driveRoot string) string {
	return filepath.Join(driveRoot, "exercise")
}

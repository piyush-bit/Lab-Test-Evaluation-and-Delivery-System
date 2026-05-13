package drive

import (
	"fmt"

	exercisestore "euc2/internals/exercise_store"
)

func (d *Drive) FetchFromDrive(id string, version string) error {
	packagePath, err := d.GetExerciseFile(id, version)
	if err != nil {
		return fmt.Errorf("resolve exercise from drive: %w", err)
	}

	if err := exercisestore.SavePackage(exercisestore.GetPublicCacheDir(), packagePath); err != nil {
		return fmt.Errorf("save exercise to local cache: %w", err)
	}
	return nil
}

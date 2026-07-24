package exercisestore

import (
	"archive/tar"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"TDES/internals/exercise"
)

const (
	indexFileName    = "index.json"
	hashShardLength  = 2
	indexFilePerm    = 0644
	artifactFilePerm = 0644
)

type IndexEntry struct {
	Latest   string            `json:"latest"`
	Versions map[string]string `json:"versions"`
}

type Index map[string]IndexEntry

func SavePackage(storeRoot string, packagePath string) error {
	manifest, err := readManifestFromPackage(packagePath)
	if err != nil {
		return err
	}
	if manifest.LabID == "" {
		return fmt.Errorf("manifest is missing lab_id")
	}
	if manifest.Version == "" {
		return fmt.Errorf("manifest is missing version")
	}

	packageHash, err := calculateFileHash(packagePath)
	if err != nil {
		return fmt.Errorf("calculate package hash: %w", err)
	}

	index, err := LoadIndex(storeRoot)
	if err != nil {
		return err
	}

	destPath, err := PackagePathFromHash(storeRoot, packageHash)
	if err != nil {
		return err
	}
	if err := copyFile(packagePath, destPath, artifactFilePerm); err != nil {
		return fmt.Errorf("copy exercise package: %w", err)
	}

	entry := index[manifest.LabID]
	if entry.Versions == nil {
		entry.Versions = make(map[string]string)
	}
	entry.Latest = manifest.Version
	entry.Versions[manifest.Version] = packageHash
	index[manifest.LabID] = entry

	if err := WriteIndex(storeRoot, index); err != nil {
		return err
	}
	return nil
}

func LoadIndex(storeRoot string) (Index, error) {
	if err := ensureRoot(storeRoot); err != nil {
		return nil, err
	}

	indexPath := indexPath(storeRoot)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			index := make(Index)
			if err := WriteIndex(storeRoot, index); err != nil {
				return nil, err
			}
			return index, nil
		}
		return nil, fmt.Errorf("read index: %w", err)
	}

	if len(data) == 0 {
		return make(Index), nil
	}

	index := make(Index)
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("decode index: %w", err)
	}
	return index, nil
}

func WriteIndex(storeRoot string, index Index) error {
	if index == nil {
		index = make(Index)
	}
	if err := ensureRoot(storeRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("encode index: %w", err)
	}

	if err := os.WriteFile(indexPath(storeRoot), data, indexFilePerm); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

func RemoveExercise(storeRoot string, exerciseID string, version string) error {
	index, err := LoadIndex(storeRoot)
	if err != nil {
		return err
	}

	entry, ok := index[exerciseID]
	if !ok {
		return nil
	}

	if version == "" {
		version = entry.Latest
	}

	packageHash, ok := entry.Versions[version]
	if ok {
		path, err := PackagePathFromHash(storeRoot, packageHash)
		if err == nil && path != "" {
			_ = os.Remove(path)
		}
		delete(entry.Versions, version)
	}

	if len(entry.Versions) == 0 {
		delete(index, exerciseID)
	} else {
		if entry.Latest == version {
			var newLatest string
			for v := range entry.Versions {
				newLatest = v
			}
			entry.Latest = newLatest
		}
		index[exerciseID] = entry
	}

	return WriteIndex(storeRoot, index)
}

func ResolveExercisePath(storeRoot string, exerciseID string, version string) (string, error) {
	index, err := LoadIndex(storeRoot)
	if err != nil {
		return "", err
	}
	return ResolveExercisePathFromIndex(storeRoot, index, exerciseID, version)
}

func ResolveExercisePathFromIndex(storeRoot string, index Index, exerciseID string, version string) (string, error) {
	entry, ok := index[exerciseID]
	if !ok {
		return "", fmt.Errorf("exercise %q not found in store", exerciseID)
	}

	if version == "" {
		version = entry.Latest
	}
	if version == "" {
		return "", fmt.Errorf("exercise %q has no stored versions", exerciseID)
	}

	packageHash, ok := entry.Versions[version]
	if !ok {
		return "", fmt.Errorf("exercise %q version %q not found in store", exerciseID, version)
	}

	path, err := PackagePathFromHash(storeRoot, packageHash)
	if err != nil {
		return "", fmt.Errorf("resolve stored package path: %w", err)
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("stored archive missing at %q", path)
		}
		return "", fmt.Errorf("stat stored archive: %w", err)
	}
	return path, nil
}

func PackagePathFromHash(storeRoot string, packageHash string) (string, error) {
	if len(packageHash) < hashShardLength {
		return "", fmt.Errorf("invalid package hash %q", packageHash)
	}
	return filepath.Join(storeRoot, packageHash[:hashShardLength], packageHash), nil
}

func ensureRoot(storeRoot string) error {
	if storeRoot == "" {
		return fmt.Errorf("store root is required")
	}
	if err := os.MkdirAll(storeRoot, 0755); err != nil {
		return fmt.Errorf("create store root: %w", err)
	}
	return nil
}

func indexPath(storeRoot string) string {
	return filepath.Join(storeRoot, indexFileName)
}

func readManifestFromPackage(packagePath string) (*exercise.ExerciseManifest, error) {
	file, err := os.Open(packagePath)
	if err != nil {
		return nil, fmt.Errorf("open exercise package: %w", err)
	}
	defer file.Close()

	tarReader := tar.NewReader(bufio.NewReader(file))
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read exercise archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != "manifest.json" {
			continue
		}

		data, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read manifest.json: %w", err)
		}

		var manifest exercise.ExerciseManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("decode manifest.json: %w", err)
		}
		return &manifest, nil
	}

	return nil, fmt.Errorf("manifest.json not found in exercise package")
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyFile(srcPath, destPath string, perm os.FileMode) error {
	if _, err := os.Stat(destPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(destDir, "exercise-store-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	if _, err := io.Copy(tempFile, srcFile); err != nil {
		return err
	}
	if err := tempFile.Chmod(perm); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		return err
	}
	return nil
}

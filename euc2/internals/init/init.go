package init

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	exercisestore "euc2/internals/exercise_store"
)

func InitFromID(id string, version string, workingDir string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("exercise id is required")
	}

	packagePath, err := exercisestore.ResolveExercisePath(exercisestore.GetPublicCacheDir(), id, version)
	if err != nil {
		return fmt.Errorf("resolve exercise package from cache: %w", err)
	}

	return Init(packagePath, workingDir)
}

func Init(sourcePath string, workingDir string) error {
	sourcePath = strings.TrimSpace(sourcePath)
	if sourcePath == "" {
		return errors.New("source package path is required")
	}
	if workingDir == "" {
		workingDir = "."
	}

	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	if err := os.MkdirAll(absWorkingDir, 0755); err != nil {
		return fmt.Errorf("create working directory: %w", err)
	}

	packageFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open exercise package: %w", err)
	}
	defer packageFile.Close()

	tarReader := tar.NewReader(packageFile)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("exercise archive not found in package")
		}
		if err != nil {
			return fmt.Errorf("read exercise package: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(filepath.Base(header.Name), ".tar.gz") {
			continue
		}

		if err := extractTarGz(tarReader, absWorkingDir); err != nil {
			return fmt.Errorf("extract %s: %w", header.Name, err)
		}
		return nil
	}
}

func extractTarGz(reader io.Reader, destinationDir string) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read archive entry: %w", err)
		}

		targetPath, err := safeTargetPath(destinationDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("create parent directory for %s: %w", targetPath, err)
			}

			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file %s: %w", targetPath, err)
			}

			_, copyErr := io.Copy(file, tarReader)
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("write file %s: %w", targetPath, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close file %s: %w", targetPath, closeErr)
			}
		default:
			return fmt.Errorf("unsupported archive entry type %d for %s", header.Typeflag, header.Name)
		}
	}
}

func safeTargetPath(destinationDir string, entryName string) (string, error) {
	cleaned := filepath.Clean(entryName)
	if cleaned == "." {
		return destinationDir, nil
	}
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive entry path: %s", entryName)
	}

	targetPath := filepath.Join(destinationDir, cleaned)
	destinationPrefix := destinationDir + string(filepath.Separator)
	if targetPath != destinationDir && !strings.HasPrefix(targetPath, destinationPrefix) {
		return "", fmt.Errorf("archive entry escapes destination directory: %s", entryName)
	}
	return targetPath, nil
}

func SplitIdWithVersion(idWithVersion string) (string, string) {
	parts := strings.Split(idWithVersion, ":")
	if len(parts) != 2 {
		return idWithVersion, ""
	}
	return parts[0], parts[1]
}

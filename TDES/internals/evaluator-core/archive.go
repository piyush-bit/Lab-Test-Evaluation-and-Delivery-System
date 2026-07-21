package evaluatorcore

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const submissionManifestFileName = "submission-manifest.json"

type SubmissionIdentity struct {
	OrgID     string `json:"org_id"`
	StudentID string `json:"student_id"`
	LabID     string `json:"lab_id"`
	Version   string `json:"version"`
}

type submissionArchiveManifest struct {
	SchemaVersion string   `json:"schema_version"`
	OrgID         string   `json:"org_id"`
	StudentID     string   `json:"student_id"`
	LabID         string   `json:"lab_id"`
	Version       string   `json:"version"`
	CreatedAt     string   `json:"created_at"`
	IncludedPaths []string `json:"included_paths"`
}

type submissionArchive struct {
	Manifest submissionArchiveManifest
	Files    map[string][]byte
}

type exerciseManifest struct {
	LabID           string         `json:"lab_id"`
	Title           string         `json:"title"`
	Version         string         `json:"version"`
	Language        string         `json:"language"`
	RunnerImage     string         `json:"runner_image"`
	LocalEntrypoint string         `json:"local_entrypoint"`
	Grading         []gradingEntry `json:"grading"`
	Submission      submissionSpec `json:"submission"`
	Limits          limitsSpec     `json:"limits"`
}

type gradingEntry struct {
	Command string `json:"command"`
	Points  int    `json:"points"`
	Public  bool   `json:"public"`
}

type submissionSpec struct {
	IncludePaths []string `json:"include_paths"`
	PrivateGlobs []string `json:"private_globs"`
	ExcludeGlobs []string `json:"exclude_globs"`
}

type limitsSpec struct {
	MemoryMB       int `json:"memory_mb"`
	TimeoutSeconds int `json:"timeout_seconds"`
	PidsLimit      int `json:"pids_limit"`
}

func readSubmissionArchive(path string) (submissionArchive, error) {
	file, err := os.Open(path)
	if err != nil {
		return submissionArchive{}, fmt.Errorf("open submission archive: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	archive := submissionArchive{Files: map[string][]byte{}}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return submissionArchive{}, fmt.Errorf("read submission archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			return submissionArchive{}, fmt.Errorf("unsupported submission entry type %d for %s", header.Typeflag, header.Name)
		}

		relPath, err := normalizeArchivePath(header.Name)
		if err != nil {
			return submissionArchive{}, err
		}

		body, err := io.ReadAll(tr)
		if err != nil {
			return submissionArchive{}, fmt.Errorf("read submission entry %s: %w", relPath, err)
		}

		if filepath.Base(relPath) == submissionManifestFileName {
			if err := json.Unmarshal(body, &archive.Manifest); err != nil {
				return submissionArchive{}, fmt.Errorf("decode submission manifest: %w", err)
			}
			continue
		}

		if _, exists := archive.Files[relPath]; exists {
			return submissionArchive{}, fmt.Errorf("duplicate submission entry: %s", relPath)
		}
		archive.Files[relPath] = body
	}

	if strings.TrimSpace(archive.Manifest.LabID) == "" {
		return submissionArchive{}, fmt.Errorf("submission manifest not found in archive")
	}

	return archive, nil
}

func extractExercisePackage(reader io.Reader, destinationDir string) error {
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
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
		return extractTarGz(tr, destinationDir)
	}
}

func extractTarGz(reader io.Reader, destinationDir string) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
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
			_, copyErr := io.Copy(file, tr)
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

func readManifest(path string) (exerciseManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return exerciseManifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()

	var manifest exerciseManifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return exerciseManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func validateManifest(manifest exerciseManifest) error {
	if strings.TrimSpace(manifest.RunnerImage) == "" {
		return fmt.Errorf("manifest is missing runner_image")
	}
	if strings.TrimSpace(manifest.LocalEntrypoint) != "make test-public" {
		return fmt.Errorf("manifest local_entrypoint must be %q", "make test-public")
	}
	if len(manifest.Grading) == 0 {
		return fmt.Errorf("manifest is missing grading")
	}
	if len(manifest.Submission.IncludePaths) == 0 {
		return fmt.Errorf("manifest is missing submission.include_paths")
	}
	for i, entry := range manifest.Grading {
		if strings.TrimSpace(entry.Command) == "" {
			return fmt.Errorf("manifest grading[%d].command is required", i)
		}
		if entry.Points < 0 {
			return fmt.Errorf("manifest grading[%d].points must be non-negative", i)
		}
	}
	return nil
}

func normalizeArchivePath(name string) (string, error) {
	cleaned := filepath.ToSlash(filepath.Clean(strings.TrimSpace(name)))
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("unsafe archive entry path: %s", name)
	}
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." || pathIsAbs(cleaned) {
		return "", fmt.Errorf("unsafe archive entry path: %s", name)
	}
	return cleaned, nil
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

func safeJoin(root string, relPath string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(relPath))
	if cleaned == "." {
		return "", fmt.Errorf("invalid relative path: %s", relPath)
	}
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe relative path: %s", relPath)
	}
	return filepath.Join(root, cleaned), nil
}

func pathIsAbs(path string) bool {
	return filepath.IsAbs(filepath.FromSlash(path))
}

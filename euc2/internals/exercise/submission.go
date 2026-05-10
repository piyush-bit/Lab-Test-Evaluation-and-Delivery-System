package exercise

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const submissionManifestFileName = "submission-manifest.json"

type SubmissionArchiveManifest struct {
	SchemaVersion string   `json:"schema_version"`
	OrgID         string   `json:"org_id"`
	StudentID     string   `json:"student_id"`
	LabID         string   `json:"lab_id"`
	CreatedAt     string   `json:"created_at"`
	IncludedPaths []string `json:"included_paths"`
}

func (e *Exercise) CreateSubmissionPackage(submissionPath *os.File, orgID string, studentID string) error {
	
	// validate inputs
	{
		if e == nil {
			return fmt.Errorf("exercise is nil")
		}
		if submissionPath == nil {
			return fmt.Errorf("submission package file is required")
		}
		if strings.TrimSpace(e.Path) == "" {
			return fmt.Errorf("exercise path is required")
		}
		orgID = strings.TrimSpace(orgID)
		if orgID == "" {
			return fmt.Errorf("org id is required")
		}
		studentID = strings.TrimSpace(studentID)
		if studentID == "" {
			return fmt.Errorf("student id is required")
		}
	
		if e.Manifest == nil {
			manifest, err := LoadManifest(e.Path)
			if err != nil {
				return fmt.Errorf("load manifest: %w", err)
			}
			e.Manifest = manifest
		}
		if len(e.Manifest.Submission.IncludePaths) == 0 {
			return fmt.Errorf("manifest is missing submission.include_paths")
		}
	}

	included, err := collectSubmissionFiles(e.Path, e.Manifest.Submission.IncludePaths)
	if err != nil {
		return err
	}
	if len(included) == 0 {
		return fmt.Errorf("no files matched submission.include_paths")
	}

	if _, err := submissionPath.Seek(0, 0); err != nil {
		return fmt.Errorf("seek submission package: %w", err)
	}
	if err := submissionPath.Truncate(0); err != nil {
		return fmt.Errorf("truncate submission package: %w", err)
	}

	tw := tar.NewWriter(submissionPath)
	submissionManifest, err := buildSubmissionArchiveManifest(e.Manifest, orgID, studentID, included)
	if err != nil {
		tw.Close()
		return err
	}
	if err := writeSubmissionManifest(tw, submissionManifest); err != nil {
		tw.Close()
		return err
	}

	for _, relPath := range included {
		sourcePath, err := safeJoin(e.Path, relPath)
		if err != nil {
			tw.Close()
			return err
		}

		info, err := os.Stat(sourcePath)
		if err != nil {
			tw.Close()
			return fmt.Errorf("stat submission file %q: %w", relPath, err)
		}
		if !info.Mode().IsRegular() {
			tw.Close()
			return fmt.Errorf("unsupported file type in %s", relPath)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			tw.Close()
			return fmt.Errorf("build tar header for %q: %w", relPath, err)
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			tw.Close()
			return fmt.Errorf("write tar header for %q: %w", relPath, err)
		}

		file, err := os.Open(sourcePath)
		if err != nil {
			tw.Close()
			return fmt.Errorf("open submission file %q: %w", relPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			file.Close()
			tw.Close()
			return fmt.Errorf("copy submission file %q: %w", relPath, err)
		}
		if err := file.Close(); err != nil {
			tw.Close()
			return fmt.Errorf("close submission file %q: %w", relPath, err)
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close submission tar writer: %w", err)
	}
	if _, err := submissionPath.Seek(0, 0); err != nil {
		return fmt.Errorf("rewind submission package: %w", err)
	}
	return nil
}

func buildSubmissionArchiveManifest(manifest *ExerciseManifest, orgID string, studentID string, includedPaths []string) (SubmissionArchiveManifest, error) {
	if manifest == nil {
		return SubmissionArchiveManifest{}, fmt.Errorf("exercise manifest is required")
	}

	return SubmissionArchiveManifest{
		SchemaVersion: "v1",
		OrgID:         orgID,
		StudentID:     studentID,
		LabID:         manifest.LabID,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		IncludedPaths: append([]string(nil), includedPaths...),
	}, nil
}

func writeSubmissionManifest(tw *tar.Writer, manifest SubmissionArchiveManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode submission manifest: %w", err)
	}
	data = append(data, '\n')

	header := &tar.Header{
		Name: submissionManifestFileName,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write submission manifest header: %w", err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("write submission manifest: %w", err)
	}
	return nil
}

func ReadSubmissionArchiveManifest(submissionPath string) (*SubmissionArchiveManifest, error) {
	file, err := os.Open(submissionPath)
	if err != nil {
		return nil, fmt.Errorf("open submission archive: %w", err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read submission archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != submissionManifestFileName {
			continue
		}

		data, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read submission manifest: %w", err)
		}

		var manifest SubmissionArchiveManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("decode submission manifest: %w", err)
		}
		return &manifest, nil
	}

	return nil, fmt.Errorf("submission manifest not found in archive")
}

func collectSubmissionFiles(exercisePath string, includePaths []string) ([]string, error) {
	seen := make(map[string]struct{})

	for _, includePath := range includePaths {
		includePath = filepath.ToSlash(strings.TrimSpace(includePath))
		if includePath == "" {
			continue
		}
		if err := validateRelativeSubmissionPath(includePath); err != nil {
			return nil, err
		}

		sourcePath, err := safeJoin(exercisePath, includePath)
		if err != nil {
			return nil, err
		}

		info, err := os.Stat(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("submission include path %q is invalid: %w", includePath, err)
		}

		if info.IsDir() {
			err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if info.IsDir() {
					return nil
				}
				if !info.Mode().IsRegular() {
					return fmt.Errorf("unsupported file type in %q", filepath.ToSlash(path))
				}

				relToRoot, err := filepath.Rel(exercisePath, path)
				if err != nil {
					return err
				}
				seen[filepath.ToSlash(relToRoot)] = struct{}{}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk submission include path %q: %w", includePath, err)
			}
			continue
		}

		seen[includePath] = struct{}{}
	}

	files := make([]string, 0, len(seen))
	for path := range seen {
		files = append(files, path)
	}
	sort.Strings(files)
	return files, nil
}


func safeJoin(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root %q: %w", root, err)
	}

	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, rel))
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", rel, err)
	}

	relToRoot, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", fmt.Errorf("check containment %q: %w", rel, err)
	}

	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes root", rel)
	}

	return targetAbs, nil
}

func validateRelativeSubmissionPath(rel string) error {
	if filepath.IsAbs(rel) {
		return fmt.Errorf("path %q escapes exercise root", rel)
	}

	for _, segment := range strings.Split(filepath.ToSlash(rel), "/") {
		if segment == ".." {
			return fmt.Errorf("path %q escapes exercise root", rel)
		}
	}

	return nil
}

package exercisestore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"TDES/internals/exercise"
)

func TestSaveAndResolvePackage(t *testing.T) {
	storeRoot := t.TempDir()
	manifest := &exercise.ExerciseManifest{
		LabID:   "go101-lab01",
		Version: "1.0.0",
	}

	packagePath := filepath.Join(t.TempDir(), "exercise.tar")
	if err := writeExercisePackage(packagePath, manifest); err != nil {
		t.Fatalf("write package: %v", err)
	}

	if err := SavePackage(storeRoot, packagePath); err != nil {
		t.Fatalf("save package: %v", err)
	}

	index, err := LoadIndex(storeRoot)
	if err != nil {
		t.Fatalf("load index: %v", err)
	}

	entry, ok := index["go101-lab01"]
	if !ok {
		t.Fatal("expected go101-lab01 in index")
	}
	if entry.Latest != manifest.Version {
		t.Fatalf("expected latest %q, got %q", manifest.Version, entry.Latest)
	}
	if got := entry.Versions[manifest.Version]; got == "" {
		t.Fatalf("expected version %q to be indexed", manifest.Version)
	}

	resolvedPath, err := ResolveExercisePath(storeRoot, manifest.LabID, manifest.Version)
	if err != nil {
		t.Fatalf("resolve package: %v", err)
	}
	if resolvedPath == "" {
		t.Fatal("expected resolved package path")
	}
	if _, err := os.Stat(resolvedPath); err != nil {
		t.Fatalf("stat resolved path: %v", err)
	}
}

func writeExercisePackage(path string, manifest *exercise.ExerciseManifest) error {
	innerArchive, err := buildInnerArchive()
	if err != nil {
		return err
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	if err := writeTarEntry(tarWriter, "manifest.json", manifestBytes); err != nil {
		return err
	}
	return writeTarEntry(tarWriter, "exercise.tar.gz", innerArchive)
}

func buildInnerArchive() ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	if err := writeTarEntry(tarWriter, "README.md", []byte("hello")); err != nil {
		return nil, err
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeTarEntry(tarWriter *tar.Writer, name string, body []byte) error {
	header := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(body)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	_, err := tarWriter.Write(body)
	return err
}

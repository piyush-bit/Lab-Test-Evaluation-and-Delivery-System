package exercise

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageExercise(t *testing.T) {
	exerciseDir := t.TempDir()
	writeManifestForTest(t, exerciseDir, ExerciseManifest{})
	writeFileForTest(t, filepath.Join(exerciseDir, "solution.go"), "package main")

	packageFile, err := os.CreateTemp(t.TempDir(), "*.tar")
	if err != nil {
		t.Fatalf("create package file: %v", err)
	}
	defer packageFile.Close()

	exercise := &Exercise{Path: exerciseDir}
	if err := BuildPackage(packageFile, exercise, nil, PackageTypePublic); err != nil {
		t.Fatalf("BuildPackage failed: %v", err)
	}

	entries := readTarEntriesForTest(t, packageFile.Name())
	if len(entries) != 2 {
		t.Fatalf("expected 2 package entries, got %d", len(entries))
	}
	if entries[0].Path != "manifest.json" {
		t.Fatalf("expected manifest.json in package, got %q", entries[0].Path)
	}
	if entries[1].Path != publicPackageArchiveName {
		t.Fatalf("expected %q in package, got %q", publicPackageArchiveName, entries[1].Path)
	}

	gzr, err := gzip.NewReader(bytes.NewReader(entries[1].Body))
	if err != nil {
		t.Fatalf("open inner gzip: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	header, err := tr.Next()
	if err != nil {
		t.Fatalf("read inner tar: %v", err)
	}
	if header.Name != "manifest.json" {
		t.Fatalf("expected inner manifest.json, got %q", header.Name)
	}
	if _, err := io.Copy(io.Discard, tr); err != nil {
		t.Fatalf("drain inner manifest: %v", err)
	}

	header, err = tr.Next()
	if err != nil {
		t.Fatalf("read inner tar file: %v", err)
	}
	if header.Name != "solution.go" {
		t.Fatalf("expected inner solution.go, got %q", header.Name)
	}
}


type tarEntry struct {
	Path string
	Body []byte
}

func readTarEntriesForTest(t *testing.T, path string) []tarEntry {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open tar: %v", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var entries []tarEntry
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read tar entry: %v", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read tar body: %v", err)
		}
		entries = append(entries, tarEntry{
			Path: filepath.ToSlash(header.Name),
			Body: body,
		})
	}
	return entries
}

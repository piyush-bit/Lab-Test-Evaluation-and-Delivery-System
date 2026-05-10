package init

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"euc2/internals/exercise"
	exercisestore "euc2/internals/exercise_store"
)

func TestInitFromIDInitializesWorkspaceFromStoredPublicPackage(t *testing.T) {
	storeDir := t.TempDir()
	t.Setenv("EUC2_CACHE_DIR", storeDir)

	createStoredPublicPackage(t, storeDir, exercise.ExerciseManifest{
		LabID:   "go101-lab01",
		Version: "1.0.0",
		Submission: exercise.Submission{
			IncludePaths: []string{"starter/main.go"},
			PrivateGlobs: []string{"tests/**", "secret.txt"},
			ExcludeGlobs: []string{"teacher-notes.txt"},
		},
	}, map[string]string{
		"README.md":            "# stacks\n",
		"starter/main.go":      "package main\n",
		"tests/private_test.go": "package main\n",
		"secret.txt":           "hidden\n",
		"teacher-notes.txt":    "omit me\n",
	})

	workingDir := filepath.Join(t.TempDir(), "workspace")
	if err := InitFromID("go101-lab01", "", workingDir); err != nil {
		t.Fatalf("InitFromID failed: %v", err)
	}

	assertFileContents(t, filepath.Join(workingDir, "README.md"), "# stacks\n")
	assertFileContents(t, filepath.Join(workingDir, "starter", "main.go"), "package main\n")
	assertManifest(t, filepath.Join(workingDir, "manifest.json"), "go101-lab01", "1.0.0")
	assertFileMissing(t, filepath.Join(workingDir, "tests", "private_test.go"))
	assertFileMissing(t, filepath.Join(workingDir, "secret.txt"))
	assertFileMissing(t, filepath.Join(workingDir, "teacher-notes.txt"))
}

func TestInitFromIDUsesLatestStoredVersionWhenVersionEmpty(t *testing.T) {
	storeDir := t.TempDir()
	t.Setenv("EUC2_CACHE_DIR", storeDir)

	createStoredPublicPackage(t, storeDir, exercise.ExerciseManifest{
		LabID:   "go101-lab01",
		Version: "1.0.0",
	}, map[string]string{
		"README.md": "version one\n",
	})
	createStoredPublicPackage(t, storeDir, exercise.ExerciseManifest{
		LabID:   "go101-lab01",
		Version: "2.0.0",
	}, map[string]string{
		"README.md": "version two\n",
	})

	workingDir := filepath.Join(t.TempDir(), "workspace")
	if err := InitFromID("go101-lab01", "", workingDir); err != nil {
		t.Fatalf("InitFromID failed: %v", err)
	}

	assertFileContents(t, filepath.Join(workingDir, "README.md"), "version two\n")
}

func TestInitRejectsUnsafeArchivePath(t *testing.T) {
	t.Parallel()

	packagePath := filepath.Join(t.TempDir(), "exercise-package.tar")
	if err := writeExercisePackage(packagePath, map[string]string{
		"../evil.txt": "bad",
	}); err != nil {
		t.Fatalf("write exercise package: %v", err)
	}

	workingDir := t.TempDir()
	err := Init(packagePath, workingDir)
	if err == nil {
		t.Fatal("expected unsafe path error")
	}
	if !strings.Contains(err.Error(), "unsafe archive entry path") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workingDir, "..", "evil.txt")); !os.IsNotExist(err) {
		t.Fatalf("unsafe file was written or stat failed: %v", err)
	}
}

func createStoredPublicPackage(t *testing.T, storeDir string, manifest exercise.ExerciseManifest, files map[string]string) {
	t.Helper()

	exerciseDir := t.TempDir()
	writeManifestForTest(t, exerciseDir, manifest)
	for path, content := range files {
		writeFileForTest(t, filepath.Join(exerciseDir, path), content)
	}

	packageFile, err := os.CreateTemp(t.TempDir(), "*.tar")
	if err != nil {
		t.Fatalf("create package file: %v", err)
	}
	defer packageFile.Close()

	exerciseFixture := &exercise.Exercise{Path: exerciseDir}
	publicExcludeGlobs := append([]string{}, manifest.Submission.ExcludeGlobs...)
	publicExcludeGlobs = append(publicExcludeGlobs, manifest.Submission.PrivateGlobs...)
	if err := exercise.BuildPackage(packageFile, exerciseFixture, publicExcludeGlobs, exercise.PackageTypePublic); err != nil {
		t.Fatalf("BuildPackage failed: %v", err)
	}
	if err := exercisestore.SavePackage(storeDir, packageFile.Name()); err != nil {
		t.Fatalf("SavePackage failed: %v", err)
	}
}

func writeManifestForTest(t *testing.T, dir string, manifest exercise.ExerciseManifest) {
	t.Helper()

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

func writeFileForTest(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("unexpected file contents for %s: got %q want %q", path, string(data), want)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be absent, stat err=%v", path, err)
	}
}

func assertManifest(t *testing.T, path, wantLabID, wantVersion string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v", path, err)
	}

	var manifest exercise.ExerciseManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal manifest %s: %v", path, err)
	}

	if manifest.LabID != wantLabID {
		t.Fatalf("unexpected manifest lab id: got %q want %q", manifest.LabID, wantLabID)
	}
	if manifest.Version != wantVersion {
		t.Fatalf("unexpected manifest version: got %q want %q", manifest.Version, wantVersion)
	}
}

func writeExercisePackage(path string, files map[string]string) error {
	innerArchive, err := buildInnerArchive(files)
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

	if err := writeTarEntry(tarWriter, "manifest.json", "{}"); err != nil {
		return err
	}

	header := &tar.Header{
		Name: "exercise.tar.gz",
		Mode: 0644,
		Size: int64(len(innerArchive)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	_, err = tarWriter.Write(innerArchive)
	return err
}

func buildInnerArchive(files map[string]string) ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)

	for name, body := range files {
		if err := writeTarEntry(tarWriter, name, body); err != nil {
			return nil, err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeTarEntry(tarWriter *tar.Writer, name string, body string) error {
	header := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(body)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	_, err := tarWriter.Write([]byte(body))
	return err
}

package exercise

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func PackageExercise(path string) (publicPath string, privatePath string, err error) {
	/**
	Steps to do this : lets go
	1. Create the exercise struct
	2. Test the exercise using the grading commands
	3. build the public and private tar (by removing the public and private globs)
	4. gz the tar files
	5. create package by creating tar of manifest.json and (public/private) gz files
	**/

	manifest, err := LoadManifest(path)
	if err != nil {
		return "", "", err
	}

	exercise := &Exercise{
		Manifest: manifest,
		Path:     path,
	}

	err = exercise.TestExercise()
	if err != nil {
		return "", "", fmt.Errorf("test exercise: %w", err)
	}
	publicExcludeGlobs := []string{}
	publicExcludeGlobs = append(publicExcludeGlobs, exercise.Manifest.Submission.ExcludeGlobs...)
	publicExcludeGlobs = append(publicExcludeGlobs, exercise.Manifest.Submission.PrivateGlobs...)

	publicPackage, err := os.CreateTemp("/tmp", "*.tar")
	if err != nil {
		return "", "", fmt.Errorf("create public package file: %w", err)
	}
	defer publicPackage.Close()
	defer func() {
		if err != nil {
			os.Remove(publicPackage.Name())
		}
	}()
	err = BuildPackage(publicPackage, exercise, publicExcludeGlobs, PackageTypePublic)
	if err != nil {
		return "", "", fmt.Errorf("build public package: %w", err)
	}

	privatePackage, err := os.CreateTemp("/tmp", "*.tar")
	if err != nil {
		return "", "", fmt.Errorf("create private package file: %w", err)
	}
	defer privatePackage.Close()
	defer func() {
		if err != nil {
			os.Remove(privatePackage.Name())
		}
	}()
	err = BuildPackage(privatePackage, exercise, exercise.Manifest.Submission.ExcludeGlobs, PackageTypePrivate)
	if err != nil {
		return "", "", fmt.Errorf("build private package: %w", err)
	}

	return publicPackage.Name(), privatePackage.Name(), nil
}

func BuildPackage(packageFile *os.File, exercise *Exercise, excludeGlob []string, pt PackageType) error {
	var innerArchiveName string
	if pt == PackageTypePublic {
		innerArchiveName = publicPackageArchiveName
	} else {
		innerArchiveName = privatePackageArchiveName
	}

	tarFile, err := os.CreateTemp("/tmp", "*.tar")
	if err != nil {
		return fmt.Errorf("create tar file: %w", err)
	}
	defer tarFile.Close()
	defer os.Remove(tarFile.Name())
	err = createTar(tarFile, exercise.Path, excludeGlob)
	if err != nil {
		return fmt.Errorf("create tar: %w", err)
	}
	compressedTarFile, err := os.CreateTemp("/tmp", "*.tar.gz")
	if err != nil {
		return fmt.Errorf("create compressed tar file: %w", err)
	}
	defer compressedTarFile.Close()
	defer os.Remove(compressedTarFile.Name())
	err = compressGz(compressedTarFile, tarFile)
	if err != nil {
		return fmt.Errorf("compress tar: %w", err)
	}
	err = createTarFromFiles(packageFile, innerArchiveName, exercise.Path+"/manifest.json", compressedTarFile.Name())
	if err != nil {
		return fmt.Errorf("create package: %w", err)
	}
	return nil
}

func createTar(destination *os.File, root string, exclude []string) error {
	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("seek destination: %w", err)
	}
	if err := destination.Truncate(0); err != nil {
		return fmt.Errorf("truncate destination: %w", err)
	}

	tw := tar.NewWriter(destination)

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if shouldExcludePath(relPath, info.IsDir(), exclude) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header for %s: %w", relPath, err)
		}
		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
			return tw.WriteHeader(header)
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", relPath)
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header for %s: %w", relPath, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file %s: %w", relPath, err)
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("copy file %s: %w", relPath, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar writer: %w", err)
	}
	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("rewind destination: %w", err)
	}
	return nil
}

// createTarFromFiles bundles the given files into a tar archive.
// innerArchiveName is the name to use inside the tar for the .tar.gz payload;
// all other files (e.g. manifest.json) keep their base filename.
func createTarFromFiles(destination *os.File, innerArchiveName string, files ...string) error {
	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("seek destination: %w", err)
	}
	if err := destination.Truncate(0); err != nil {
		return fmt.Errorf("truncate destination: %w", err)
	}

	tw := tar.NewWriter(destination)

	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("stat file %s: %w", filePath, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", filePath)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("create tar header for %s: %w", filePath, err)
		}
		header.Name = archiveNameForPackageFile(filePath, innerArchiveName)

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("write tar header for %s: %w", filePath, err)
		}

		if err := copyFileToTar(tw, filePath); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar writer: %w", err)
	}
	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("rewind destination: %w", err)
	}
	return nil
}

func copyFileToTar(tw *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer file.Close()

	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("copy file %s: %w", filePath, err)
	}
	return nil
}

func compressGz(destination, source *os.File) error {
	if _, err := source.Seek(0, 0); err != nil {
		return fmt.Errorf("seek source: %w", err)
	}
	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("seek destination: %w", err)
	}
	if err := destination.Truncate(0); err != nil {
		return fmt.Errorf("truncate destination: %w", err)
	}

	gw := gzip.NewWriter(destination)
	if _, err := io.Copy(gw, source); err != nil {
		gw.Close()
		return fmt.Errorf("compress gzip: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("close gzip writer: %w", err)
	}

	if _, err := destination.Seek(0, 0); err != nil {
		return fmt.Errorf("rewind destination: %w", err)
	}
	return nil
}

func shouldExcludePath(path string, isDir bool, exclude []string) bool {
	for _, pattern := range exclude {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		if matched, err := filepath.Match(pattern, path); err == nil && matched {
			return true
		}

		if strings.HasSuffix(pattern, "/*") {
			dirPattern := strings.TrimSuffix(pattern, "/*")
			if path == dirPattern || strings.HasPrefix(path, dirPattern+"/") {
				return true
			}
		}

		if isDir {
			dirPattern := strings.TrimSuffix(pattern, "/")
			if path == dirPattern || strings.HasPrefix(path, dirPattern+"/") {
				return true
			}
		}
	}
	return false
}

// archiveNameForPackageFile returns the name to store for filePath inside the
// outer package tar. The .tar.gz payload is stored under innerArchiveName
// (which carries the public/private prefix); everything else keeps its
// base filename.
func archiveNameForPackageFile(filePath, innerArchiveName string) string {
	base := filepath.Base(filePath)
	if strings.HasSuffix(base, ".tar.gz") {
		return innerArchiveName
	}
	return base
}

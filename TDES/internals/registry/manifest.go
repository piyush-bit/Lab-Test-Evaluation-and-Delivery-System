package registry

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"TDES/internals/exercise"
)

// ReadManifestFromPackage extracts and decodes manifest.json from a package archive (tar or tar.gz).
func ReadManifestFromPackage(packagePath string) (*exercise.ExerciseManifest, error) {
	file, err := os.Open(packagePath)
	if err != nil {
		return nil, fmt.Errorf("open package: %w", err)
	}
	defer file.Close()

	br := bufio.NewReader(file)
	// Peek at the first 2 bytes to check for gzip magic number (0x1f 0x8b)
	headerBytes, err := br.Peek(2)
	
	var reader io.Reader = br
	if err == nil && len(headerBytes) >= 2 && headerBytes[0] == 0x1f && headerBytes[1] == 0x8b {
		gzipReader, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read package: %w", err)
		}

		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != "manifest.json" {
			continue
		}

		data, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read manifest: %w", err)
		}

		var manifest exercise.ExerciseManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return nil, fmt.Errorf("decode manifest: %w", err)
		}
		return &manifest, nil
	}

	return nil, fmt.Errorf("manifest.json not found in package")
}

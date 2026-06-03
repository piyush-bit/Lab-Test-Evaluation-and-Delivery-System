package exercisestore

import (
	"os"
	"path/filepath"
	"strings"
)

const cacheDirEnvVar = "EUC2_CACHE_DIR"

func GetPublicCacheDir() string {
	if cacheDir := strings.TrimSpace(os.Getenv(cacheDirEnvVar)); cacheDir != "" {
		return cacheDir
	}

	userCacheDir, err := os.UserCacheDir()
	if err == nil && userCacheDir != "" {
		return filepath.Join(userCacheDir, "euc2", "cache")
	}

	return filepath.Join(os.TempDir(), "euc2", "cache")
}

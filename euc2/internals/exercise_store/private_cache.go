package exercisestore

import (
	"os"
	"path/filepath"
	"strings"
)

const storeDirEnvVar = "EUC2_PRIVATE_STORE_DIR"

func GetPrivateCacheDir() string {
	if storeDir := strings.TrimSpace(os.Getenv(storeDirEnvVar)); storeDir != "" {
		return storeDir
	}

	userConfigDir, err := os.UserConfigDir()
	if err == nil && userConfigDir != "" {
		return filepath.Join(userConfigDir, "euc2", "private_packages")
	}

	return filepath.Join(os.TempDir(), "euc2", "private_packages")
}

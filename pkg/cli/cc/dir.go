package cc

import (
	"os"
	"path/filepath"
	"runtime"
)

// Common topaz directory paths and operations.

// GetTopazDir returns the topaz root directory ($TOPAZ_DIR).
func GetTopazDir() string {
	if topazDir := os.Getenv("TOPAZ_DIR"); topazDir != "" {
		return topazDir
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(filepath.Join(getHomeFallback(), ".config", "topaz"))
	}
	return filepath.Clean(filepath.Join(homeDir, ".config", "topaz"))
}

// GetTopazCfgDir returns the topaz config directory ($TOPAZ_DIR/cfg).
func GetTopazCfgDir() string {
	if cfgDir := os.Getenv("TOPAZ_CFG_DIR"); cfgDir != "" {
		return cfgDir
	}
	return filepath.Clean(filepath.Join(GetTopazDir(), "cfg"))
}

// GetTopazCertsDir returns the topaz certs directory ($TOPAZ_DIR/certs).
func GetTopazCertsDir() string {
	if certsDir := os.Getenv("TOPAZ_CERTS_DIR"); certsDir != "" {
		return certsDir
	}
	return filepath.Clean(filepath.Join(GetTopazDir(), "certs"))
}

// GetTopazDataDir returns the topaz db directory ($TOPAZ_DIR/db).
func GetTopazDataDir() string {
	if dataDir := os.Getenv("TOPAZ_DB_DIR"); dataDir != "" {
		return dataDir
	}
	return filepath.Clean(filepath.Join(GetTopazDir(), "db"))
}

const (
	fallback        = `~/.config/topaz`
	darwinFallBack  = fallback
	linuxFallBack   = fallback
	windowsFallBack = `\\.config\\topaz`
)

func getHomeFallback() string {
	switch runtime.GOOS {
	case "darwin":
		return darwinFallBack
	case "linux":
		return linuxFallBack
	case "windows":
		return filepath.Join(os.Getenv("USERPROFILE"), windowsFallBack)
	default:
		return ""
	}
}

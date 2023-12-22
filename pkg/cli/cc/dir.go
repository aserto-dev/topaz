package cc

import (
	"os"
	"path/filepath"
	"runtime"
)

// Common topaz directory paths and operations.

// GetTopazDir returns the topaz root directory ($TOPAZ_DIR).
func GetTopazDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Clean(filepath.Join(getHomeFallback(), ".config", "topaz"))
	}
	return filepath.Clean(filepath.Join(homeDir, ".config", "topaz"))
}

// GetTopazCfgDir returns the topaz config directory ($TOPAZ_DIR/cfg).
func GetTopazCfgDir() string {
	return filepath.Clean(filepath.Join(GetTopazDir(), "cfg"))
}

// GetTopazCertsDir returns the topaz certs directory ($TOPAZ_DIR/certs).
func GetTopazCertsDir() string {
	return filepath.Clean(filepath.Join(GetTopazDir(), "certs"))
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

// FileName returns the filename part of a fully qualified file path.
func FileName(fqn string) string {
	_, fn := filepath.Split(fqn)
	return fn
}

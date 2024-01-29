package cc

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// Common topaz directory paths and operations.

// GetTopazDir returns the topaz root directory ($TOPAZ_DIR).
func GetTopazDir() string {
	if topazDir := os.Getenv("TOPAZ_DIR"); topazDir != "" {
		return topazDir
	}
	return filepath.Clean(filepath.Join(xdg.Home, ".config", "topaz"))
}

// GetTopazCfgDir returns the topaz config directory ($TOPAZ_DIR/cfg).
func GetTopazCfgDir() string {
	if cfgDir := os.Getenv("TOPAZ_CFG_DIR"); cfgDir != "" {
		return cfgDir
	}
	return filepath.Clean(filepath.Join(xdg.ConfigHome, "topaz", "cfg"))
}

// GetTopazCertsDir returns the topaz certs directory ($TOPAZ_DIR/certs).
func GetTopazCertsDir() string {
	if certsDir := os.Getenv("TOPAZ_CERTS_DIR"); certsDir != "" {
		return certsDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "certs"))
}

// GetTopazDataDir returns the topaz db directory ($TOPAZ_DIR/db).
func GetTopazDataDir() string {
	if dataDir := os.Getenv("TOPAZ_DB_DIR"); dataDir != "" {
		return dataDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "db"))
}

package cc

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
	"github.com/aserto-dev/topaz/pkg/fs"
)

// Common topaz directory paths and operations.

// GetTopazDir returns the topaz root directory ($HOME/.config/topaz).
func GetTopazDir() string {
	if topazDir := os.Getenv("TOPAZ_DIR"); topazDir != "" {
		return topazDir
	}
	return filepath.Clean(filepath.Join(xdg.ConfigHome, "topaz"))
}

// GetTopazCfgDir returns the topaz config directory ($XDG_CONFIG_HOME/topaz/cfg).
func GetTopazCfgDir() string {
	if cfgDir := os.Getenv("TOPAZ_CFG_DIR"); cfgDir != "" {
		return cfgDir
	}
	return filepath.Clean(filepath.Join(xdg.ConfigHome, "topaz", "cfg"))
}

// GetTopazCertsDir returns the topaz certs directory ($XDG_DATA_HOME/topaz/certs).
func GetTopazCertsDir() string {
	if certsDir := os.Getenv("TOPAZ_CERTS_DIR"); certsDir != "" {
		return certsDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "certs"))
}

// GetTopazDataDir returns the topaz db directory ($XDG_DATA_HOME/topaz/db).
func GetTopazDataDir() string {
	if dataDir := os.Getenv("TOPAZ_DB_DIR"); dataDir != "" {
		return dataDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "db"))
}

// GetTopazTemplateDir returns the templates installation directory ($XDG_DATA_HOME/topaz/tmpl).
func GetTopazTemplateDir() string {
	if tmplDir := os.Getenv("TOPAZ_TMPL_DIR"); tmplDir != "" {
		return tmplDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "tmpl"))
}

func EnsureDirs() error {
	for _, f := range []func() error{EnsureTopazDir, EnsureTopazCfgDir, EnsureTopazCertsDir, EnsureTopazDataDir, EnsureTopazTemplateDir} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func EnsureTopazDir() error {
	return fs.EnsureDirPath(GetTopazDir(), fs.FileMode0700)
}

func EnsureTopazCfgDir() error {
	return fs.EnsureDirPath(GetTopazCfgDir(), fs.FileMode0700)
}

func EnsureTopazCertsDir() error {
	return fs.EnsureDirPath(GetTopazCertsDir(), fs.FileMode0755)
}

func EnsureTopazDataDir() error {
	return fs.EnsureDirPath(GetTopazDataDir(), fs.FileMode0700)
}

func EnsureTopazTemplateDir() error {
	return fs.EnsureDirPath(GetTopazTemplateDir(), fs.FileMode0700)
}

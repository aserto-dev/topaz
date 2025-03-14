package cc

import (
	"os"
	"path/filepath"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
	"github.com/aserto-dev/topaz/pkg/cli/x"
)

// Common topaz directory paths and operations.

// GetTopazDir returns the topaz root directory ($HOME/.config/topaz).
func GetTopazDir() string {
	if topazDir := os.Getenv(x.EnvTopazDir); topazDir != "" {
		return topazDir
	}
	return filepath.Clean(filepath.Join(xdg.ConfigHome, "topaz"))
}

// GetTopazCfgDir returns the topaz config directory ($XDG_CONFIG_HOME/topaz/cfg).
func GetTopazCfgDir() string {
	if cfgDir := os.Getenv(x.EnvTopazCfgDir); cfgDir != "" {
		return cfgDir
	}
	return filepath.Clean(filepath.Join(xdg.ConfigHome, "topaz", "cfg"))
}

// GetTopazCertsDir returns the topaz certs directory ($XDG_DATA_HOME/topaz/certs).
func GetTopazCertsDir() string {
	if certsDir := os.Getenv(x.EnvTopazCertsDir); certsDir != "" {
		return certsDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "certs"))
}

// GetTopazDataDir returns the topaz db directory ($XDG_DATA_HOME/topaz/db).
func GetTopazDataDir() string {
	if dataDir := os.Getenv(x.EnvTopazDBDir); dataDir != "" {
		return dataDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "db"))
}

// GetTopazTemplateDir returns the templates installation directory ($XDG_DATA_HOME/topaz/tmpl).
func GetTopazTemplateDir() string {
	if tmplDir := os.Getenv(x.EnvTopazTmplDir); tmplDir != "" {
		return tmplDir
	}
	return filepath.Clean(filepath.Join(xdg.DataHome, "topaz", "tmpl"))
}

// GetTopazTemplateURL returns the URL to the templates container, can be local or remote.
func GetTopazTemplateURL() string {
	if tmplURL := os.Getenv(x.EnvTopazTmplURL); tmplURL != "" {
		return tmplURL
	}
	return x.DefTopazTmplURL
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
	dir := GetTopazDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

func EnsureTopazCfgDir() error {
	dir := GetTopazCfgDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

func EnsureTopazCertsDir() error {
	dir := GetTopazCertsDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func EnsureTopazDataDir() error {
	dir := GetTopazDataDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

func EnsureTopazTemplateDir() error {
	dir := GetTopazTemplateDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dir, 0o700)
}

package testing

import (
	"path/filepath"
	"runtime"

	"github.com/aserto-dev/topaz/pkg/cc/config"
)

// AssetsDir returns the directory containing test assets
// nolint: dogsled
func AssetsDir() string {
	_, filename, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(filename), "assets")
}

// AssetAcmeEBBFilePath returns the path of the test contoso EDS database file.
func AssetAcmeEBBFilePath() string {
	return filepath.Join(AssetsDir(), "eds-acmecorp.db")
}

// AssetDefaultConfigOnline returns the path of the default yaml config file that uses an online bundle.
func AssetDefaultConfigOnline() config.Path {
	return config.Path(filepath.Join(AssetsDir(), "config-online.yaml"))
}

// AssetDefaultConfigLocal returns the path of the default yaml config file that doesn't use an online bundle.
func AssetDefaultConfigLocal() config.Path {
	return config.Path(filepath.Join(AssetsDir(), "config-local.yaml"))
}

// AssetLocalBundle returns the path of the default local bundle directory
// that contains test policies and data.
func AssetLocalBundle() string {
	return filepath.Join(AssetsDir(), "mycars_bundle")
}

package testutil

import (
	"path/filepath"
	"runtime"
)

// AssetsDir returns the directory containing test assets.
func AssetsDir() string {
	_, filename, _, _ := runtime.Caller(0) //nolint:dogsled

	return filepath.Join(filepath.Dir(filename), "assets")
}

// AssetMycarsBundle returns the path of the mycars local bundle directory
// that contains test policies and data.
func AssetMycarsBundle() string {
	return filepath.Join(AssetsDir(), "mycars_bundle")
}

// AssetFakeBuiltinsBundle returns the path of a bundle that uses a fake builtin.
func AssetFakeBuiltinsBundle() string {
	return filepath.Join(AssetsDir(), "fake_builtin")
}

// AssetSimpleBundle returns the path of a bundle that contains one rego file with one rule.
func AssetSimpleBundle() string {
	return filepath.Join(AssetsDir(), "simple")
}

// AssetBuiltinsBundle returns the path of a bundle that uses a builtin.
func AssetBuiltinsBundle() string {
	return filepath.Join(AssetsDir(), "builtin")
}

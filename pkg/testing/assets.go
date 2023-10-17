package testing

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/pkg/errors"
)

// AssetsDir returns the directory containing test assets
// nolint: dogsled
func AssetsDir() string {
	_, filename, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(filename), "assets")
}

// AssetAcmeEBBFilePath returns the path of the test contoso EDS database file.
func AssetAcmeEBBFilePath() string {
	const filename = "eds-acmecorp.db"
	srcFile := filepath.Join(AssetsDir(), filename)
	dstFile := filepath.Join(os.TempDir(), filename)

	if err := fcopy(srcFile, dstFile, true); err != nil {
		panic(err)
	}

	return dstFile
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

// fcopy file copy with conditional overwrite.
func fcopy(src, dst string, overwrite bool) error {
	const bufferSize = int64(1024)

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return errors.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if !overwrite {
		if _, err = os.Stat(dst); err == nil {
			return errors.Wrapf(err, "file %s already exists", dst)
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return err
}

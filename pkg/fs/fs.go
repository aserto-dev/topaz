package fs

import (
	"os"

	"github.com/pkg/errors"
)

const (
	FileMode0644 os.FileMode = 0o644
	FileMode0700 os.FileMode = 0o700
	FileMode0755 os.FileMode = 0o755
)

func FileExists(path string) bool {
	if fsInfo, err := os.Stat(path); err == nil && !fsInfo.IsDir() {
		return true
	}
	return false
}

func FileExistsWithErr(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func DirExists(path string) bool {
	if fsInfo, err := os.Stat(path); err == nil && fsInfo.IsDir() {
		return true
	}
	return false
}

func DirExistsWithErr(path string) (bool, error) {
	fsInfo, err := os.Stat(path)
	if err == nil && fsInfo.IsDir() {
		return true, nil
	}
	return false, err
}

func EnsureDirPath(path string, perm os.FileMode) error {
	if !DirExists(path) {
		return os.MkdirAll(path, perm)
	}
	return nil
}

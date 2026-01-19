package fs

import (
	"os"

	"github.com/pkg/errors"
)

const (
	FileModeOwnerRO      os.FileMode = 0o400 // dr--------
	FileModeOwnerRW      os.FileMode = 0o600 // drw-------
	FileModeOwnerRWX     os.FileMode = 0o700 // drwx------
	FileModeDirectoryRWX os.FileMode = 0o755 // drwxr-xr-x
)

func FileExists(path string) bool {
	if fsInfo, err := os.Stat(path); err == nil && !fsInfo.IsDir() {
		return true
	}

	return false
}

func FileExistsEx(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func DirExists(path string) bool {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return true
	}

	return false
}

func DirExistsEx(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

func EnsureDirPath(path string, perm os.FileMode) error {
	if !DirExists(path) {
		return os.MkdirAll(path, perm)
	}

	return nil
}

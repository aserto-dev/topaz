package fs

import (
	"os"
)

const (
	FileModeOwnerRW  os.FileMode = 0o600
	FileModeOwnerRWX os.FileMode = 0o700
)

func FileExists(path string) bool {
	fsInfo, err := os.Stat(path)
	if err == nil && !fsInfo.IsDir() {
		return true
	}

	return false
}

func DirExists(path string) bool {
	fsInfo, err := os.Stat(path)
	if err == nil && fsInfo.IsDir() {
		return true
	}

	return false
}

func EnsureDirPath(path string) error {
	if !DirExists(path) {
		return os.MkdirAll(path, FileModeOwnerRWX)
	}

	return nil
}

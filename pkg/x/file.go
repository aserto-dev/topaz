package x

import (
	"os"

	"github.com/pkg/errors"
)

func FileExists(path string) (bool, error) {
	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

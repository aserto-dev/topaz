package fflag

import (
	"os"
	"strconv"
	"sync"

	"github.com/aserto-dev/topaz/topaz/x"
)

// feature flags package.
const (
	Default FFlag = 0
)

type FFlag uint64

const (
	Editor FFlag = 1 << iota
	Prompter
)

var (
	ffOnce sync.Once
	ff     FFlag
)

func Init() {
	ffOnce.Do(func() {
		env := os.Getenv(x.EnvTopazFeatureFlag)
		if env == "" {
			ff = Default
		}

		f, err := strconv.ParseUint(os.Getenv(x.EnvTopazFeatureFlag), 10, 8)
		if err != nil {
			ff = Default
		}

		ff = FFlag(f)
	})
}

func FF() FFlag {
	return ff
}

func Enabled(flag FFlag) bool {
	return ff&flag != 0
}

func (f FFlag) IsSet(flag FFlag) bool {
	return f&flag != 0
}

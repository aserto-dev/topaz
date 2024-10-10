package fflag

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// feature flags package.
const (
	Env     string = "TOPAZ_FFLAG"
	Default FFlag  = 0
)

type FFlag uint64

const (
	Editor FFlag = 1 << iota
	Prompter
)

var flags = map[uint64]string{
	uint64(Editor):   "editor",
	uint64(Prompter): "prompter",
}

var (
	ffOnce sync.Once
	ff     FFlag
)

func Init() {
	ffOnce.Do(func() {
		env := os.Getenv(Env)
		if env == "" {
			ff = Default
		}

		f, err := strconv.ParseUint(os.Getenv(Env), 10, 8)
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

func F(s string) FFlag {
	for k, v := range flags {
		if v == s {
			return FFlag(k)
		}
	}
	return Default
}

func (f FFlag) IsSet(flag FFlag) bool {
	return f&flag != 0
}

func (f FFlag) Base() uint64 {
	return uint64(f)
}

func (f FFlag) String() string {
	str := []string{}
	for k, v := range flags {
		if f.IsSet(FFlag(k)) {
			str = append(str, v)
		}
	}
	return strings.Join(str, "|")
}

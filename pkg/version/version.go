package version

import (
	"fmt"
	"runtime"
	"time"

	"github.com/aserto-dev/topaz/pkg/cli/x"
)

var (
	ver    string //nolint:gochecknoglobals // set by linker
	date   string //nolint:gochecknoglobals // set by linker
	commit string //nolint:gochecknoglobals // set by linker
)

// Info - version info.
type Info struct {
	Version string
	Date    string
	Commit  string
}

// GetInfo - get version stamp information.
func GetInfo() Info {
	if ver == "" {
		ver = "0.0.0-dev"
	}

	if date == "" {
		date = time.Now().Format(time.RFC3339)
	}

	if commit == "" {
		commit = "undefined"
	}

	return Info{
		Version: ver,
		Date:    date,
		Commit:  commit,
	}
}

// String() -- return version info string.
func (vi Info) String() string {
	return fmt.Sprintf("%s g%s %s-%s [%s]",
		vi.Version,
		vi.Commit,
		runtime.GOOS,
		runtime.GOARCH,
		vi.Date,
	)
}

func UserAgent() string {
	return fmt.Sprintf("%s/%s", x.AppName, GetInfo().Version)
}

//go:build aix || dragonfly || freebsd || (js && wasm) || nacl || linux || netbsd || openbsd || solaris

package xdg_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
)

func TestDefaultBaseDirs(t *testing.T) {
	home := xdg.Home

	testDirs(t,
		&envSample{
			name:     "XDG_DATA_HOME",
			expected: filepath.Join(home, ".local/share"),
			actual:   &xdg.DataHome,
		},
		&envSample{
			name:     "XDG_DATA_DIRS",
			expected: []string{"/usr/local/share", "/usr/share"},
			actual:   &xdg.DataDirs,
		},
		&envSample{
			name:     "XDG_CONFIG_HOME",
			expected: filepath.Join(home, ".config"),
			actual:   &xdg.ConfigHome,
		},
		&envSample{
			name:     "XDG_CONFIG_DIRS",
			expected: []string{"/etc/xdg"},
			actual:   &xdg.ConfigDirs,
		},
		&envSample{
			name:     "XDG_STATE_HOME",
			expected: filepath.Join(home, ".local", "state"),
			actual:   &xdg.StateHome,
		},
		&envSample{
			name:     "XDG_CACHE_HOME",
			expected: filepath.Join(home, ".cache"),
			actual:   &xdg.CacheHome,
		},
		&envSample{
			name:     "XDG_RUNTIME_DIR",
			expected: filepath.Join("/run/user", strconv.Itoa(os.Getuid())),
			actual:   &xdg.RuntimeDir,
		},
		&envSample{
			name: "XDG_APPLICATION_DIRS",
			expected: []string{
				filepath.Join(home, ".local/share/applications"),
				"/usr/local/share/applications",
				"/usr/share/applications",
			},
			actual: &xdg.ApplicationDirs,
		},
		&envSample{
			name: "XDG_FONT_DIRS",
			expected: []string{
				filepath.Join(home, ".local/share/fonts"),
				filepath.Join(home, ".fonts"),
				"/usr/local/share/fonts",
				"/usr/share/fonts",
			},
			actual: &xdg.FontDirs,
		},
	)
}

func TestCustomBaseDirs(t *testing.T) {
	home := xdg.Home

	testDirs(t,
		&envSample{
			name:     "XDG_DATA_HOME",
			value:    "~/.local/data",
			expected: filepath.Join(home, ".local/data"),
			actual:   &xdg.DataHome,
		},
		&envSample{
			name:     "XDG_DATA_DIRS",
			value:    "~/.local/data:/usr/share",
			expected: []string{filepath.Join(home, ".local/data"), "/usr/share"},
			actual:   &xdg.DataDirs,
		},
		&envSample{
			name:     "XDG_CONFIG_HOME",
			value:    "~/.local/config",
			expected: filepath.Join(home, ".local/config"),
			actual:   &xdg.ConfigHome,
		},
		&envSample{
			name:     "XDG_CONFIG_DIRS",
			value:    "~/.local/config:/etc/xdg",
			expected: []string{filepath.Join(home, ".local/config"), "/etc/xdg"},
			actual:   &xdg.ConfigDirs,
		},
		&envSample{
			name:     "XDG_STATE_HOME",
			value:    "~/.local/var",
			expected: filepath.Join(home, ".local/var"),
			actual:   &xdg.StateHome,
		},
		&envSample{
			name:     "XDG_CACHE_HOME",
			value:    "~/.local/cache",
			expected: filepath.Join(home, ".local/cache"),
			actual:   &xdg.CacheHome,
		},
		&envSample{
			name:     "XDG_RUNTIME_DIR",
			value:    "~/.local/runtime",
			expected: filepath.Join(home, ".local/runtime"),
			actual:   &xdg.RuntimeDir,
		},
	)
}

func TestDefaultUserDirs(t *testing.T) {
	home := xdg.Home

	testDirs(t,
		&envSample{
			name:     "XDG_DESKTOP_DIR",
			expected: filepath.Join(home, "Desktop"),
			actual:   &xdg.UserDirs.Desktop,
		},
		&envSample{
			name:     "XDG_DOWNLOAD_DIR",
			expected: filepath.Join(home, "Downloads"),
			actual:   &xdg.UserDirs.Download,
		},
		&envSample{
			name:     "XDG_DOCUMENTS_DIR",
			expected: filepath.Join(home, "Documents"),
			actual:   &xdg.UserDirs.Documents,
		},
		&envSample{
			name:     "XDG_MUSIC_DIR",
			expected: filepath.Join(home, "Music"),
			actual:   &xdg.UserDirs.Music,
		},
		&envSample{
			name:     "XDG_PICTURES_DIR",
			expected: filepath.Join(home, "Pictures"),
			actual:   &xdg.UserDirs.Pictures,
		},
		&envSample{
			name:     "XDG_VIDEOS_DIR",
			expected: filepath.Join(home, "Videos"),
			actual:   &xdg.UserDirs.Videos,
		},
		&envSample{
			name:     "XDG_TEMPLATES_DIR",
			expected: filepath.Join(home, "Templates"),
			actual:   &xdg.UserDirs.Templates,
		},
		&envSample{
			name:     "XDG_PUBLICSHARE_DIR",
			expected: filepath.Join(home, "Public"),
			actual:   &xdg.UserDirs.PublicShare,
		},
	)
}

func TestCustomUserDirs(t *testing.T) {
	home := xdg.Home

	testDirs(t,
		&envSample{
			name:     "XDG_DESKTOP_DIR",
			value:    "$HOME/.local/Desktop",
			expected: filepath.Join(home, ".local/Desktop"),
			actual:   &xdg.UserDirs.Desktop,
		},
		&envSample{
			name:     "XDG_DOWNLOAD_DIR",
			value:    "$HOME/.local/Downloads",
			expected: filepath.Join(home, ".local/Downloads"),
			actual:   &xdg.UserDirs.Download,
		},
		&envSample{
			name:     "XDG_DOCUMENTS_DIR",
			value:    "$HOME/.local/Documents",
			expected: filepath.Join(home, ".local/Documents"),
			actual:   &xdg.UserDirs.Documents,
		},
		&envSample{
			name:     "XDG_MUSIC_DIR",
			value:    "$HOME/.local/Music",
			expected: filepath.Join(home, ".local/Music"),
			actual:   &xdg.UserDirs.Music,
		},
		&envSample{
			name:     "XDG_PICTURES_DIR",
			value:    "$HOME/.local/Pictures",
			expected: filepath.Join(home, ".local/Pictures"),
			actual:   &xdg.UserDirs.Pictures,
		},
		&envSample{
			name:     "XDG_VIDEOS_DIR",
			value:    "$HOME/.local/Videos",
			expected: filepath.Join(home, ".local/Videos"),
			actual:   &xdg.UserDirs.Videos,
		},
		&envSample{
			name:     "XDG_TEMPLATES_DIR",
			value:    "$HOME/.local/Templates",
			expected: filepath.Join(home, ".local/Templates"),
			actual:   &xdg.UserDirs.Templates,
		},
		&envSample{
			name:     "XDG_PUBLICSHARE_DIR",
			value:    "$HOME/.local/Public",
			expected: filepath.Join(home, ".local/Public"),
			actual:   &xdg.UserDirs.PublicShare,
		},
	)
}

func TestHomeNotSet(t *testing.T) {
	envHomeVar := "HOME"
	envHomeVal := os.Getenv(envHomeVar)
	require.NoError(t, os.Unsetenv(envHomeVar))

	xdg.Reload()
	require.Equal(t, "/", xdg.Home)

	require.NoError(t, os.Setenv(envHomeVar, envHomeVal))
	xdg.Reload()
	require.Equal(t, envHomeVal, xdg.Home)
}

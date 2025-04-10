//go:build windows

package xdg_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
	"github.com/stretchr/testify/require"
)

func TestDefaultBaseDirs(t *testing.T) {
	home := xdg.Home
	systemDrive := `C:\`
	roamingAppData := filepath.Join(home, "AppData", "Roaming")
	localAppData := filepath.Join(home, "AppData", "Local")
	systemRoot := filepath.Join(systemDrive, "Windows")
	programData := filepath.Join(systemDrive, "ProgramData")

	envSamples := []*envSample{
		{
			name:     "XDG_DATA_HOME",
			expected: localAppData,
			actual:   &xdg.DataHome,
		},
		{
			name:     "XDG_DATA_DIRS",
			expected: []string{roamingAppData, programData},
			actual:   &xdg.DataDirs,
		},
		{
			name:     "XDG_CONFIG_HOME",
			expected: localAppData,
			actual:   &xdg.ConfigHome,
		},
		{
			name:     "XDG_CONFIG_DIRS",
			expected: []string{programData, roamingAppData},
			actual:   &xdg.ConfigDirs,
		},
		{
			name:     "XDG_STATE_HOME",
			expected: localAppData,
			actual:   &xdg.StateHome,
		},
		{
			name:     "XDG_CACHE_HOME",
			expected: filepath.Join(localAppData, "cache"),
			actual:   &xdg.CacheHome,
		},
		{
			name:     "XDG_RUNTIME_DIR",
			expected: localAppData,
			actual:   &xdg.RuntimeDir,
		},
		{
			name: "XDG_APPLICATION_DIRS",
			expected: []string{
				filepath.Join(roamingAppData, "Microsoft", "Windows", "Start Menu", "Programs"),
				filepath.Join(programData, "Microsoft", "Windows", "Start Menu", "Programs"),
			},
			actual: &xdg.ApplicationDirs,
		},
		{
			name: "XDG_FONT_DIRS",
			expected: []string{
				filepath.Join(systemRoot, "Fonts"),
				filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"),
			},
			actual: &xdg.FontDirs,
		},
	}

	// Test default environment.
	testDirs(t, envSamples...)

	// Test system drive not set.
	envSystemDrive := os.Getenv("SystemDrive")
	require.NoError(t, os.Unsetenv("SystemDrive"))
	testDirs(t, envSamples...)
	require.NoError(t, os.Setenv("SystemDrive", envSystemDrive))
}

func TestCustomBaseDirs(t *testing.T) {
	home := xdg.Home
	roamingAppData := filepath.Join(home, "Custom", "Appdata", "Roaming")
	localAppData := filepath.Join(home, "Custom", "AppData", "Local")
	programData := filepath.Join(home, "Custom", "ProgramData")

	envRoamingAppData := os.Getenv("APPDATA")
	require.NoError(t, os.Setenv("APPDATA", roamingAppData))
	envLocalAppData := os.Getenv("LOCALAPPDATA")
	require.NoError(t, os.Setenv("LOCALAPPDATA", localAppData))
	envProgramData := os.Getenv("ProgramData")
	require.NoError(t, os.Setenv("ProgramData", programData))

	testDirs(t,
		&envSample{
			name:     "XDG_DATA_HOME",
			value:    filepath.Join(localAppData, "Data"),
			expected: filepath.Join(localAppData, "Data"),
			actual:   &xdg.DataHome,
		},
		&envSample{
			name:     "XDG_DATA_DIRS",
			value:    fmt.Sprintf("%s;%s", filepath.Join(localAppData, "Data"), filepath.Join(roamingAppData, "Data")),
			expected: []string{filepath.Join(localAppData, "Data"), filepath.Join(roamingAppData, "Data")},
			actual:   &xdg.DataDirs,
		},
		&envSample{
			name:     "XDG_CONFIG_HOME",
			value:    filepath.Join(localAppData, "Config"),
			expected: filepath.Join(localAppData, "Config"),
			actual:   &xdg.ConfigHome,
		},
		&envSample{
			name:     "XDG_CONFIG_DIRS",
			value:    fmt.Sprintf("%s;%s", filepath.Join(localAppData, "Config"), filepath.Join(roamingAppData, "Config")),
			expected: []string{filepath.Join(localAppData, "Config"), filepath.Join(roamingAppData, "Config")},
			actual:   &xdg.ConfigDirs,
		},
		&envSample{
			name:     "XDG_STATE_HOME",
			value:    filepath.Join(programData, "State"),
			expected: filepath.Join(programData, "State"),
			actual:   &xdg.StateHome,
		},
		&envSample{
			name:     "XDG_CACHE_HOME",
			value:    filepath.Join(programData, "Cache"),
			expected: filepath.Join(programData, "Cache"),
			actual:   &xdg.CacheHome,
		},
		&envSample{
			name:     "XDG_RUNTIME_DIR",
			value:    filepath.Join(programData, "Runtime"),
			expected: filepath.Join(programData, "Runtime"),
			actual:   &xdg.RuntimeDir,
		},
	)

	require.NoError(t, os.Setenv("APPDATA", envRoamingAppData))
	require.NoError(t, os.Setenv("LOCALAPPDATA", envLocalAppData))
	require.NoError(t, os.Setenv("ProgramData", envProgramData))
}

func TestDefaultUserDirs(t *testing.T) {
	home := xdg.Home
	roamingAppData := filepath.Join(home, "AppData", "Roaming")

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
			expected: filepath.Join(roamingAppData, "Microsoft", "Windows", "Templates"),
			actual:   &xdg.UserDirs.Templates,
		},
		&envSample{
			name:     "XDG_PUBLICSHARE_DIR",
			expected: filepath.Join(`C:\`, "Users", "Public"),
			actual:   &xdg.UserDirs.PublicShare,
		},
	)
}

func TestCustomUserDirs(t *testing.T) {
	home := xdg.Home

	testDirs(t,
		&envSample{
			name:     "XDG_DESKTOP_DIR",
			value:    filepath.Join(home, "Files/Desktop"),
			expected: filepath.Join(home, "Files/Desktop"),
			actual:   &xdg.UserDirs.Desktop,
		},
		&envSample{
			name:     "XDG_DOWNLOAD_DIR",
			value:    filepath.Join(home, "Files/Downloads"),
			expected: filepath.Join(home, "Files/Downloads"),
			actual:   &xdg.UserDirs.Download,
		},
		&envSample{
			name:     "XDG_DOCUMENTS_DIR",
			value:    filepath.Join(home, "Files/Documents"),
			expected: filepath.Join(home, "Files/Documents"),
			actual:   &xdg.UserDirs.Documents,
		},
		&envSample{
			name:     "XDG_MUSIC_DIR",
			value:    filepath.Join(home, "Files/Music"),
			expected: filepath.Join(home, "Files/Music"),
			actual:   &xdg.UserDirs.Music,
		},
		&envSample{
			name:     "XDG_PICTURES_DIR",
			value:    filepath.Join(home, "Files/Pictures"),
			expected: filepath.Join(home, "Files/Pictures"),
			actual:   &xdg.UserDirs.Pictures,
		},
		&envSample{
			name:     "XDG_VIDEOS_DIR",
			value:    filepath.Join(home, "Files/Videos"),
			expected: filepath.Join(home, "Files/Videos"),
			actual:   &xdg.UserDirs.Videos,
		},
		&envSample{
			name:     "XDG_TEMPLATES_DIR",
			value:    filepath.Join(home, "Files/Templates"),
			expected: filepath.Join(home, "Files/Templates"),
			actual:   &xdg.UserDirs.Templates,
		},
		&envSample{
			name:     "XDG_PUBLICSHARE_DIR",
			value:    filepath.Join(home, "Files/Public"),
			expected: filepath.Join(home, "Files/Public"),
			actual:   &xdg.UserDirs.PublicShare,
		},
	)
}

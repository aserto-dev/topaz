package xdg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
)

type envSample struct {
	name     string
	value    string
	expected interface{}
	actual   interface{}
}

func testDirs(t *testing.T, samples ...*envSample) {
	// Test home directory.
	require.NotEmpty(t, xdg.Home)
	t.Logf("Home: %s", xdg.Home)

	// Set environment variables.
	for _, sample := range samples {
		require.NoError(t, os.Setenv(sample.name, sample.value))
	}
	xdg.Reload()

	// Test results.
	for _, sample := range samples {
		var actual interface{}
		switch v := sample.actual.(type) {
		case *string:
			actual = *v
		case *[]string:
			actual = *v
		}

		require.Equal(t, sample.expected, actual)
		t.Logf("%s: %v", sample.name, actual)
	}
}

type testInputData struct {
	relPaths   []string
	pathFunc   func(string) (string, error)
	searchFunc func(string) (string, error)
}

func TestBaseDirFuncs(t *testing.T) {
	inputs := []*testInputData{
		{
			relPaths:   []string{"app.data", "appname/app.data"},
			pathFunc:   xdg.DataFile,
			searchFunc: xdg.SearchDataFile,
		},
		{
			relPaths:   []string{"app.yaml", "appname/app.yaml"},
			pathFunc:   xdg.ConfigFile,
			searchFunc: xdg.SearchConfigFile,
		},
		{
			relPaths:   []string{"app.state", "appname/app.state"},
			pathFunc:   xdg.StateFile,
			searchFunc: xdg.SearchStateFile,
		},
		{
			relPaths:   []string{"app.cache", "appname/app.cache"},
			pathFunc:   xdg.CacheFile,
			searchFunc: xdg.SearchCacheFile,
		},
		{
			relPaths:   []string{"app.pid", "appname/app.pid"},
			pathFunc:   xdg.RuntimeFile,
			searchFunc: xdg.SearchRuntimeFile,
		},
	}

	// Test base directories for regular files.
	testBaseDirsRegular(t, inputs)

	// Test base directories for symbolic links.
	for _, input := range inputs {
		input.relPaths = []string{input.relPaths[1]}
	}

	testBaseDirsSymlinks(t, inputs)
}

func testBaseDirsRegular(t *testing.T, inputs []*testInputData) {
	for _, input := range inputs {
		for _, relPath := range input.relPaths {
			// Get suitable path for input file.
			expFullPath, err := input.pathFunc(relPath)
			require.NoError(t, err)

			// Create input file.
			f, err := os.Create(expFullPath)
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Search input file after creation.
			actFullPath, err := input.searchFunc(relPath)
			require.NoError(t, err)
			require.Equal(t, expFullPath, actFullPath)

			// Remove created file.
			require.NoError(t, os.Remove(expFullPath))

			// Search input file after removal.
			_, err = input.searchFunc(relPath)
			require.Error(t, err)

			// Check that the same path is returned.
			actFullPath, err = input.pathFunc(relPath)
			require.NoError(t, err)
			require.Equal(t, expFullPath, actFullPath)
		}
	}
}

func testBaseDirsSymlinks(t *testing.T, inputs []*testInputData) {
	for _, input := range inputs {
		for _, relPath := range input.relPaths {
			// Get suitable path for input file.
			expFullPath, err := input.pathFunc(relPath)
			require.NoError(t, err)

			// Replace input directory with symlink.
			symlinkDir := filepath.Dir(expFullPath)
			inputDir := filepath.Join(filepath.Dir(symlinkDir), "inputdir")

			require.NoError(t, os.Remove(symlinkDir))
			require.NoError(t, os.Mkdir(inputDir, os.ModeDir|0o700))
			require.NoError(t, os.Symlink(inputDir, symlinkDir))

			// Create input file.
			inputPath := filepath.Join(symlinkDir, "input.file")

			f, err := os.Create(inputPath)
			require.NoError(t, err)
			require.NoError(t, f.Close())

			// Create symbolic link.
			require.NoError(t, os.Symlink(inputPath, expFullPath))

			// Search input file after creation.
			actFullPath, err := input.searchFunc(relPath)
			require.NoError(t, err)
			require.Equal(t, expFullPath, actFullPath)

			// Remove created symbolic links, files and directories.
			require.NoError(t, os.Remove(expFullPath))
			require.NoError(t, os.Remove(inputPath))
			require.NoError(t, os.Remove(symlinkDir))
			require.NoError(t, os.Remove(inputDir))

			// Search input file after removal.
			_, err = input.searchFunc(relPath)
			require.Error(t, err)

			// Check that the same path is returned.
			actFullPath, err = input.pathFunc(relPath)
			require.NoError(t, err)
			require.Equal(t, expFullPath, actFullPath)
		}
	}
}

func TestInvalidPaths(t *testing.T) {
	inputs := map[string]func(string) (string, error){
		"\000/app.data":          xdg.DataFile,
		"appname\000/app.yaml":   xdg.ConfigFile,
		"appname/\000/app.state": xdg.StateFile,
		"\000appname/app.cache":  xdg.CacheFile,
		"\000/appname/app.pid":   xdg.RuntimeFile,
	}

	for inputPath, xdgFunc := range inputs {
		_, err := xdgFunc(inputPath)
		require.Error(t, err)
	}
}

//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || nacl || linux || netbsd || openbsd || solaris

package pathutil_test

import (
	"path/filepath"
	"testing"

	"github.com/aserto-dev/topaz/internal/pkg/xdg/internal/pathutil"

	"github.com/stretchr/testify/require"
)

func TestExpandHome(t *testing.T) {
	home := "/home/test"

	require.Equal(t, home, pathutil.ExpandHome("~", home))
	require.Equal(t, home, pathutil.ExpandHome("$HOME", home))
	require.Equal(t, filepath.Join(home, "appname"), pathutil.ExpandHome("~/appname", home))
	require.Equal(t, filepath.Join(home, "appname"), pathutil.ExpandHome("$HOME/appname", home))

	require.Equal(t, "", pathutil.ExpandHome("", home))
	require.Equal(t, home, pathutil.ExpandHome(home, ""))
	require.Equal(t, "", pathutil.ExpandHome("", ""))

	require.Equal(t, home, pathutil.ExpandHome(home, home))
	require.Equal(t, "/", pathutil.ExpandHome("~", "/"))
	require.Equal(t, "/", pathutil.ExpandHome("$HOME", "/"))
	require.Equal(t, "/usr/bin", pathutil.ExpandHome("~/bin", "/usr"))
	require.Equal(t, "/usr/bin", pathutil.ExpandHome("$HOME/bin", "/usr"))
}

func TestUnique(t *testing.T) {
	input := []string{
		"",
		"/home",
		"/home/test",
		"a",
		"~/appname",
		"$HOME/appname",
		"a",
		"/home",
	}

	expected := []string{
		"/home",
		"/home/test",
		"/home/test/appname",
	}

	require.EqualValues(t, expected, pathutil.Unique(input, "/home/test"))
}

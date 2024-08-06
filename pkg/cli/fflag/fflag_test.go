package fflag_test

import (
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/stretchr/testify/assert"
)

func TestFFlag(t *testing.T) {
	{
		ff := fflag.FFlag(0)
		assert.Equal(t, false, ff.IsSet(fflag.Editor))
		assert.Equal(t, false, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(1)
		assert.Equal(t, true, ff.IsSet(fflag.Editor))
		assert.Equal(t, false, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(2)
		assert.Equal(t, false, ff.IsSet(fflag.Editor))
		assert.Equal(t, true, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(3)
		assert.Equal(t, true, ff.IsSet(fflag.Editor))
		assert.Equal(t, true, ff.IsSet(fflag.Prompter))
	}

	{
		ff := fflag.FFlag(31)
		assert.Equal(t, true, ff.IsSet(fflag.Editor))
		assert.Equal(t, true, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(63)
		assert.Equal(t, true, ff.IsSet(fflag.Editor))
		assert.Equal(t, true, ff.IsSet(fflag.Prompter))
	}
}

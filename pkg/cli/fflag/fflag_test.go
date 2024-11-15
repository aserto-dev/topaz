package fflag_test

import (
	"testing"

	"github.com/aserto-dev/topaz/pkg/cli/fflag"
	"github.com/stretchr/testify/assert"
)

func TestFFlag(t *testing.T) {
	{
		ff := fflag.FFlag(0)
		assert.False(t, ff.IsSet(fflag.Editor))
		assert.False(t, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(1)
		assert.True(t, ff.IsSet(fflag.Editor))
		assert.False(t, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(2)
		assert.False(t, ff.IsSet(fflag.Editor))
		assert.True(t, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(3)
		assert.True(t, ff.IsSet(fflag.Editor))
		assert.True(t, ff.IsSet(fflag.Prompter))
	}

	{
		ff := fflag.FFlag(31)
		assert.True(t, ff.IsSet(fflag.Editor))
		assert.True(t, ff.IsSet(fflag.Prompter))
	}
	{
		ff := fflag.FFlag(63)
		assert.True(t, ff.IsSet(fflag.Editor))
		assert.True(t, ff.IsSet(fflag.Prompter))
	}
}

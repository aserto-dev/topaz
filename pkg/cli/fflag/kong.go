package fflag

import (
	"strings"

	"github.com/alecthomas/kong"
)

func UnHideCmds(ctx *kong.Context) {
	n := ctx.Selected()
	if n == nil {
		return
	}

	for _, c := range n.Children {
		if strings.HasPrefix(c.Tag.Type, "fflag") && c.Tag.Hidden && FF().IsSet(Editor) {
			c.Hidden = false
		}
	}
}

func UnHideFlags(ctx *kong.Context) {
	n := ctx.Selected()
	if n == nil {
		return
	}
	for _, f := range n.Flags {
		if strings.HasPrefix(f.Tag.Type, "fflag") && f.Tag.Hidden && FF().IsSet(Editor) {
			f.Hidden = false
		}
	}
}

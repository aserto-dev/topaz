package cc

import (
	"context"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/topaz/pkg/cli/cc/iostream"
)

type CommonCtx struct {
	Context context.Context
	UI      *clui.UI
	NoCheck bool
}

func NewCommonContext(noCheck bool) (*CommonCtx, error) {
	return &CommonCtx{
		Context: context.Background(),
		UI:      iostream.NewUI(iostream.DefaultIO()),
		NoCheck: noCheck,
	}, nil
}

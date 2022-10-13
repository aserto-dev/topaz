package cc

import (
	"context"

	"github.com/aserto-dev/clui"
)

type CommonCtx struct {
	Context context.Context
	UI      *clui.UI
}

func NewCommonContext() (*CommonCtx, error) {
	return &CommonCtx{
		Context: context.Background(),
	}, nil
}

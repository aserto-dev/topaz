package nop

import (
	"context"

	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog"
	"github.com/rs/zerolog"
)

type nopLogger struct{}

var _ decisionlog.DecisionLogger = (*nopLogger)(nil)

func New(ctx context.Context, logger *zerolog.Logger) (*nopLogger, error) {
	return &nopLogger{}, nil
}

func (*nopLogger) Log(d *api.Decision) error {
	return nil
}

func (*nopLogger) Shutdown() {
}

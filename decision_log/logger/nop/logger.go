package nop

import (
	"context"

	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/rs/zerolog"
)

type nopLogger struct{}

func New(ctx context.Context, logger *zerolog.Logger) (decisionlog.DecisionLogger, error) {
	return &nopLogger{}, nil
}

func (*nopLogger) Log(d *api.Decision) error {
	return nil
}

func (*nopLogger) Shutdown() {
}

package nop

import (
	"context"

	"github.com/aserto-dev/aserto-grpc/grpcclient"
	decisionlog "github.com/aserto-dev/authorizer/decision_log"
	dl "github.com/aserto-dev/go-grpc/aserto/decision_logs/v1"
	"github.com/rs/zerolog"
)

const TypeName = ""

type nopLogger struct{}

func New(context.Context, map[string]interface{}, *zerolog.Logger, grpcclient.DialOptionsProvider) (decisionlog.DecisionLogger, error) {
	return &nopLogger{}, nil
}

func (*nopLogger) Log(d *dl.Decision) error {
	return nil
}

func (*nopLogger) Shutdown() {
}

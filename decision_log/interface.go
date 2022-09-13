package decisionlog

import (
	dl "github.com/aserto-dev/go-grpc/aserto/decision_logs/v1"
)

type DecisionLogger interface {
	Log(*dl.Decision) error
	Shutdown()
}

package decisionlog

import (
	api "github.com/aserto-dev/go-authorizer/aserto/api/v2"
)

type DecisionLogger interface {
	Log(*api.Decision) error
	Shutdown()
}

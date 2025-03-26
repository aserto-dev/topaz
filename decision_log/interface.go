package decisionlog

import (
	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

type DecisionLogger interface {
	Log(decision *api.Decision) error
	Shutdown()
}

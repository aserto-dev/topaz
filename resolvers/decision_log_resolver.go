package resolvers

import (
	"context"

	dl "github.com/aserto-dev/go-grpc/aserto/decision_logs/v1"
)

type DecisionLogFunc func(*dl.Decision) error

type DecisionLogResolver interface {
	DecisionLogFromContext(ctx context.Context, policyID, policyName, instanceLabel string) (DecisionLogFunc, error)
	GetDecisionLog(ctx context.Context, tenantID, policyID, policyName, instanceLabel string) (DecisionLogFunc, error)
}

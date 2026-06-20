package runtime

import (
	"context"

	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/open-policy-agent/opa/v1/storage"
	"github.com/pkg/errors"
)

func (r *Runtime) ListPolicies(ctx context.Context) ([]types.PolicyV1, error) {
	policies := []types.PolicyV1{}
	err := storage.Txn(ctx, r.pluginsManager.Store, storage.TransactionParams{}, func(txn storage.Transaction) error {
		compiler := r.pluginsManager.GetCompiler()

		ids, err := r.storage.ListPolicies(ctx, txn)
		if err != nil {
			return errors.Wrap(err, "failed to list policies")
		}

		for _, id := range ids {
			policyBs, err := r.storage.GetPolicy(ctx, txn, id)
			if err != nil {
				return errors.Wrapf(err, "failed to get policy with ID [%s]", id)
			}

			policy := types.PolicyV1{
				ID:  id,
				Raw: string(policyBs),
				AST: compiler.Modules[id],
			}

			policies = append(policies, policy)
		}

		return nil
	})

	return policies, err
}

func (r *Runtime) GetPolicy(ctx context.Context, id string) (*types.PolicyV1, error) {
	policies, err := r.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		if policy.ID == id {
			return &policy, nil
		}
	}

	return nil, err
}

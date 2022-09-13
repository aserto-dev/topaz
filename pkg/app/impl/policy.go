package impl

import (
	"context"
	"fmt"

	api "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	"github.com/aserto-dev/go-utils/cerr"
	"github.com/aserto-dev/topaz/resolvers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// PolicyServer implements a Policy Server for the GRPC API
type PolicyServer struct {
	logger          *zerolog.Logger
	runtimeResolver resolvers.RuntimeResolver
}

// NewPolicyServer creates a new PoliciesServer
func NewPolicyServer(logger *zerolog.Logger, runtimeResolver resolvers.RuntimeResolver) *PolicyServer {
	newLogger := logger.With().Str("component", "api.policy-server").Logger()
	return &PolicyServer{
		logger:          &newLogger,
		runtimeResolver: runtimeResolver,
	}
}

// ListPolicies, returns list of bundles loaded in runtime. (RUNTIME STATE)
func (p *PolicyServer) ListPolicies(ctx context.Context, req *api.ListPoliciesRequest) (*api.ListPoliciesResponse, error) {
	runtimes, err := p.runtimeResolver.ListRuntimes(ctx)
	if err != nil {
		return &api.ListPoliciesResponse{}, errors.Wrap(err, "failed to procure tenant runtime")
	}
	var result []*api.PolicyItem
	for _, runtime := range runtimes {
		bundleList, err := runtime.GetBundles(ctx)
		if err != nil {
			return &api.ListPoliciesResponse{}, errors.Wrapf(err, "get bundles")
		}
		result = append(result, bundleList...)
	}

	return &api.ListPoliciesResponse{
		Results: result,
	}, nil
}

// GetPolicies, returns list of policies for a given policy id.
func (p *PolicyServer) GetPolicies(ctx context.Context, req *api.GetPoliciesRequest) (*api.GetPoliciesResponse, error) {
	resp := &api.GetPoliciesResponse{}

	if req.GetId() == "" {
		return resp, cerr.ErrInvalidPolicyID.Msg("policy id req parameter not set")
	}

	runtime, err := p.runtimeResolver.RuntimeFromContext(ctx, req.GetId(), req.GetName(), req.InstanceLabel)
	if err != nil {
		return resp, errors.Wrap(err, "failed to procure tenant runtime")
	}

	bundle, err := runtime.GetBundleByID(ctx, req.GetId())
	if err != nil || bundle == nil {
		return resp, errors.Wrapf(err, "get bundle by id [%s]", req.GetId())
	}

	policies, err := runtime.GetPolicies(ctx, req.GetId())
	if err != nil {
		return resp, errors.Wrapf(err, "get policy list [%s]", req.GetId())
	}

	return &api.GetPoliciesResponse{
		Id:       req.GetId(),
		Name:     bundle.Name,
		Policies: policies,
	}, nil
}

// GetModule, return policy module for given module id.
func (p *PolicyServer) GetModule(ctx context.Context, req *api.GetModuleRequest) (*api.GetModuleResponse, error) {
	resp := &api.GetModuleResponse{}

	if req.Id == "" {
		return resp, cerr.ErrInvalidArgument.Msg("module id req parameter not set")
	}

	if req.GetPolicyId() == "" {
		return resp, cerr.ErrInvalidPolicyID.Msg("policy id req parameter not set")
	}

	rt, err := p.runtimeResolver.RuntimeFromContext(ctx, req.GetPolicyId(), req.GetPolicyName(), req.InstanceLabel)
	if err != nil {
		return resp, errors.Wrap(err, "failed to procure tenant runtime")
	}

	module, err := rt.GetModule(ctx, req.Id)
	if err != nil {
		msg := fmt.Sprintf("get module [%s]", req.Id)
		return resp, errors.Wrapf(err, msg)
	}

	return &api.GetModuleResponse{
		Module: module,
	}, err
}

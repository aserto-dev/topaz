package impl

import (
	"context"
	"encoding/json"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"

	"github.com/mennanov/fmutils"
	"github.com/open-policy-agent/opa/v1/server/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *AuthorizerServer) ListPolicies(ctx context.Context, req *authorizer.ListPoliciesRequest) (*authorizer.ListPoliciesResponse, error) {
	response := &authorizer.ListPoliciesResponse{}

	rt, err := s.getRuntime(ctx, s.instanceName(req))
	if err != nil {
		return response, errors.Wrap(err, "failed to get runtime")
	}

	policies, err := rt.ListPolicies(ctx)
	if err != nil {
		return response, err
	}

	for _, policy := range policies {
		module, err := policyToModule(policy)
		if err != nil {
			return response, errors.Wrapf(err, "failed to parse policy with ID [%s]", policy.ID)
		}

		if req.GetFieldMask() != nil {
			paths := s.validateMask(req.GetFieldMask(), &api.Module{})
			mask := fmutils.NestedMaskFromPaths(paths)
			mask.Filter(module)
		}

		response.Result = append(response.Result, module)
	}

	return response, nil
}

func (s *AuthorizerServer) GetPolicy(ctx context.Context, req *authorizer.GetPolicyRequest) (*authorizer.GetPolicyResponse, error) {
	response := &authorizer.GetPolicyResponse{}

	rt, err := s.getRuntime(ctx, s.instanceName(req))
	if err != nil {
		return response, errors.Wrap(err, "failed to get runtime")
	}

	policy, err := rt.GetPolicy(ctx, req.GetId())
	if err != nil {
		return response, errors.Wrapf(err, "failed to get policy with ID [%s]", req.GetId())
	}

	if policy == nil {
		return response, errors.Wrapf(aerr.ErrPolicyNotFound, "with ID [%s]", req.GetId())
	}

	module, err := policyToModule(*policy)
	if err != nil {
		return response, errors.Wrap(err, "failed to convert policy to api.module")
	}

	if req.GetFieldMask() != nil {
		paths := s.validateMask(req.GetFieldMask(), &api.Module{})
		mask := fmutils.NestedMaskFromPaths(paths)
		mask.Filter(module)
	}

	response.Result = module

	return response, nil
}

func policyToModule(policy types.PolicyV1) (*api.Module, error) {
	astBts, err := json.Marshal(policy.AST)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal AST")
	}

	var v any
	if err := json.Unmarshal(astBts, &v); err != nil {
		return nil, errors.Wrap(err, "failed to determine AST")
	}

	astValue, err := structpb.NewValue(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create astValue")
	}

	var packageName string
	if policy.AST != nil {
		packageName = policy.AST.Package.Path.String()
	}

	module := api.Module{
		Id:          &policy.ID,
		Raw:         &policy.Raw,
		PackagePath: &packageName,
		Ast:         astValue,
	}

	return &module, nil
}

// validateMask checks if provided mask is validate.
func (s *AuthorizerServer) validateMask(mask *fieldmaskpb.FieldMask, protomsg protoreflect.ProtoMessage) []string {
	if len(mask.GetPaths()) > 0 && mask.GetPaths()[0] == "" {
		return []string{}
	}

	mask.Normalize()

	if !mask.IsValid(protomsg) {
		s.logger.Error().Msgf("field mask invalid %q", mask.GetPaths())
		return []string{}
	}

	return mask.GetPaths()
}

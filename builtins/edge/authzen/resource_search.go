package authzen

import (
	"github.com/authzen/access.go/api/access/v1"

	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/topaz/builtins"
	"github.com/aserto-dev/topaz/resolvers"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"

	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

/*
RegisterCheck - az.evaluation
az.evaluation({
	"subject": {
		"type": "",
		"id": "",
		"properties": {}
	},
	"action": {
		"name": "",
		"properties": {}
	},
	"resource": {
		"type": "",
		"id": "",
		"properties": {}
	},
	"context": {}
})
*/

func RegisterResourceSearch(logger *zerolog.Logger, fnName string, dr resolvers.DirectoryResolver) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args access.ResourceSearchRequest

			if err := ast.As(op1.Value, &args); err != nil {
				return nil, err
			}

			if proto.Equal(&args, &access.ResourceSearchRequest{}) {
				return builtins.HelpMsg(fnName, resourceSearchHelp())
			}

			resp, err := dr.GetAuthZen().ResourceSearch(bctx.Context, &args)
			if err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			return builtins.ResponseToTerm(resp)
		}
}

func resourceSearchHelp() *access.ResourceSearchRequest {
	return &access.ResourceSearchRequest{
		Subject: &access.Subject{
			Type:       "",
			Id:         "",
			Properties: pb.NewStruct(),
		},
		Action: &access.Action{
			Name:       "",
			Properties: pb.NewStruct(),
		},
		Resource: &access.Resource{
			Type:       "",
			Id:         "",
			Properties: pb.NewStruct(),
		},
		Context: pb.NewStruct(),
		Page: &access.Page{
			NextToken: "",
		},
	}
}

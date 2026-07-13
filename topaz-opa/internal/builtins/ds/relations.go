package ds

import (
	"bytes"
	"errors"

	"github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins"
	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

const dsRelationsHelp string = `ds.relations({
	object_type: "",
	object_id: "",
	relation: "",
	subject_type: "",
	subject_id: "",
	subject_relation: "",
	with_objects: false,
	with_empty_subject_relation: false
})`

// registerRelations - ds.relations.
func registerRelations(fnName string, dr func() (reader.ReaderClient, error)) (*rego.Function, rego.Builtin1) {
	return &rego.Function{
			Name:    fnName,
			Decl:    types.NewFunction(types.Args(types.A), types.A),
			Memoize: true,
		},
		func(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
			var args reader.GetRelationsRequest

			dsr, err := dr()
			if err != nil && errors.Is(err, errs.ErrTopazPluginDisabled) {
				return nil, err
			}

			if err := ast.As(op1.Value, &args); err != nil {
				builtins.TraceError(&bctx, fnName, err)
				return nil, err
			}

			if proto.Equal(&args, &reader.GetRelationsRequest{}) {
				return ast.StringTerm(dsRelationsHelp), nil
			}

			args.Page = &common.PaginationRequest{Size: builtins.MaxPaginationSize, Token: ""}

			resp := &reader.GetRelationsResponse{}

			for {
				r, err := dsr.GetRelations(bctx.Context, &args)
				if err != nil {
					builtins.TraceError(&bctx, fnName, err)
					return nil, err
				}

				resp.Results = append(resp.GetResults(), r.GetResults()...)
				resp.Objects = lo.Assign(resp.GetObjects(), r.GetObjects())

				if r.GetPage().GetNextToken() == "" {
					break
				}

				args.Page.Token = r.GetPage().GetNextToken()
			}

			buf := new(bytes.Buffer)

			var result proto.Message
			if resp.Results != nil {
				result = resp
			}

			if err := builtins.ProtoToBuf(buf, result); err != nil {
				return nil, err
			}

			v, err := ast.ValueFromReader(buf)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		}
}

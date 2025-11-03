package builtins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoToBuf, marshal proto message to buffer.
func ProtoToBuf(w io.Writer, msg proto.Message) error {
	b, err := protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "",
		AllowPartial:      false,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   false,
		EmitDefaultValues: true,
	}.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

// TraceError.
func TraceError(bctx *topdown.BuiltinContext, fnName string, err error) {
	if bctx.TraceEnabled {
		if len(bctx.QueryTracers) > 0 {
			bctx.QueryTracers[0].TraceEvent(topdown.Event{
				Op:      topdown.FailOp,
				Message: fmt.Sprintf("%s error:%s", fnName, err.Error()),
			})
		}
	}
}

type Message[T any] interface {
	proto.Message
	*T
}

// ResponseToTerm.
func ResponseToTerm[T any, M Message[T]](resp M) (*ast.Term, error) {
	buf := new(bytes.Buffer)
	if err := ProtoToBuf(buf, resp); err != nil {
		return nil, err
	}

	result := map[string]any{}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}

	v, err := ast.InterfaceToValue(result)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(v), nil
}

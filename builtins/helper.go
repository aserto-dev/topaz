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

func Help(fnName string, args any) (*ast.Term, error) {
	m := map[string]any{fnName: args}

	val, err := ast.InterfaceToValue(m)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}

func HelpMsg(fnName string, msg proto.Message) (*ast.Term, error) {
	v, err := ProtoToInterface(msg)
	if err != nil {
		return nil, err
	}

	m := map[string]any{fnName: v}

	val, err := ast.InterfaceToValue(m)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(val), nil
}

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

// BufToProto, unmarshal buffer to proto message instance.
func BufToProto(r io.Reader, msg proto.Message) error {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	return protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(buf.Bytes(), msg)
}

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

func ProtoToInterface(msg proto.Message) (any, error) {
	b, err := protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}.Marshal(msg)
	if err != nil {
		return nil, err
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	return v, nil
}

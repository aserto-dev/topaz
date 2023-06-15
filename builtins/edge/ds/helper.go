package ds

import (
	"bytes"
	"fmt"
	"io"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func help(fnName string, args interface{}) (*ast.Term, error) {
	m := map[string]interface{}{fnName: args}
	val, err := ast.InterfaceToValue(m)
	if err != nil {
		return nil, err
	}
	return ast.NewTerm(val), nil
}

// ProtoToBuf, marshal proto message to buffer.
func ProtoToBuf(w io.Writer, msg proto.Message) error {
	b, err := protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
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

func traceError(bctx *topdown.BuiltinContext, fnName string, err error) {
	if bctx.TraceEnabled {
		if len(bctx.QueryTracers) > 0 {
			bctx.QueryTracers[0].TraceEvent(topdown.Event{
				Op:      topdown.FailOp,
				Message: fmt.Sprintf("%s error:%s", fnName, err.Error()),
			})
		}
	}
}

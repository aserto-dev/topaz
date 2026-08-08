package grpcc

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoToAny, explicitly converts proto message to `any` in order
// to preserve enum values as strings when marshaled to JSON.
// NOTE: returns nil to indicate failure!
func ProtoToAny(msg proto.Message) any {
	b, err := protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "  ",
		AllowPartial:      false,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   true,
		EmitDefaultValues: false,
	}.Marshal(msg)
	if err != nil {
		return nil
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}

	return v
}

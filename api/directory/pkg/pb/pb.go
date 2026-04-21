package pb

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

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

// UnmarshalNext, JSON decoder helper function to unmarshal next message.
func UnmarshalNext(d *json.Decoder, m proto.Message) error {
	var b json.RawMessage
	if err := d.Decode(&b); err != nil {
		return err
	}

	return protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
	}.Unmarshal(b, m)
}

// ProtoToStr, marshal proto message to string representation.
func ProtoToStr(msg proto.Message) string {
	return protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "  ",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}.Format(msg)
}

// NewStruct, returns *structpb.Struct instance with initialized Fields map.
func NewStruct() *structpb.Struct {
	return &structpb.Struct{Fields: make(map[string]*structpb.Value)}
}

// ProtoToBytes, marshal proto message to buffer.
func ProtoToBytes(msg proto.Message) ([]byte, error) {
	return protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    false,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: false,
	}.Marshal(msg)
}

func BytesToProto(b []byte, msg proto.Message) error {
	return protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(b, msg)
}

func BytesToStruct(b []byte) (*structpb.Struct, error) {
	v := NewStruct()
	if len(b) == 0 {
		return v, nil
	}

	err := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(b, v)
	if err != nil {
		return NewStruct(), err
	}

	return v, nil
}

// JSONToStruct converts a map decoded from JSON to a protobuf struct.
// The reason that the map can't be directly converted to a struct using structpb.NewStruct is that
// when gqlgen decodes a JSON object it calls decoder.UseNumber() on the json.Decoder.
// As a result, numeric values are decoded as json.Number, instead of float, and structpb.NewStruct doesn't know
// how to handle that.
func JSONToStruct(val map[string]any) (*structpb.Struct, error) {
	encoded, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	pb := structpb.Struct{}
	if err := pb.UnmarshalJSON(encoded); err != nil {
		return nil, err
	}

	return &pb, nil
}

// Contains returns true if the collection contains the message.
func Contains[T proto.Message](collection []T, message T) bool {
	return lo.ContainsBy(collection, func(item T) bool {
		return proto.Equal(item, message)
	})
}

package jsonx

import (
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func OutputJSONPB(w io.Writer, msg proto.Message, opts ...protojson.MarshalOptions) error {
	options := DefaultMarshalOpts()
	if len(opts) == 1 {
		options = opts[0]
	}

	b, err := options.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

func DefaultMarshalOpts() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "  ",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   true,
		EmitDefaultValues: false,
	}
}

func MaskedMarshalOpts() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "  ",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   false,
		EmitDefaultValues: false,
	}
}

func MarshalOpts(multiline bool) protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:         multiline,
		Indent:            "  ",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   true,
		EmitDefaultValues: false,
	}
}

func LineMarshalOpts() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:         false,
		Indent:            "",
		AllowPartial:      true,
		UseProtoNames:     true,
		UseEnumNumbers:    false,
		EmitUnpopulated:   true,
		EmitDefaultValues: false,
	}
}

type Encoder struct {
	w  io.Writer
	mo protojson.MarshalOptions
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:  w,
		mo: LineMarshalOpts(),
	}
}

func (e *Encoder) Encode(m proto.Message) error {
	b, err := e.mo.Marshal(m)
	if err != nil {
		return err
	}

	_, err = e.w.Write(b)
	if err != nil {
		return err
	}

	_, err = e.w.Write([]byte("\n"))

	return err
}

func Unmarshal(b []byte, m proto.Message) error {
	return protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: false,
		RecursionLimit: 0,
	}.Unmarshal(b, m)
}

package js

import (
	"fmt"
	"os"

	"github.com/aserto-dev/go-directory/pkg/pb"
	"google.golang.org/protobuf/proto"
)

type Writer struct {
	w     *os.File
	first bool
}

func NewWriter(path, key string) (*Writer, error) {
	w, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	f := Writer{
		w:     w,
		first: false,
	}

	_, _ = f.w.WriteString("{\n")
	_, _ = f.w.WriteString(fmt.Sprintf("%q:\n", key)) //nolint: gocritic
	_, _ = f.w.WriteString("[\n")

	return &f, nil
}

func (f *Writer) Close() error {
	if f.w != nil {
		_, _ = f.w.WriteString("]\n")
		_, _ = f.w.WriteString("}\n")
		f.first = false
		err := f.w.Close()
		f.w = nil
		return err
	}
	return nil
}

func (f *Writer) Write(msg proto.Message) error {
	if f.first {
		_, _ = f.w.WriteString(",")
	}

	err := pb.ProtoToBuf(f.w, msg)

	if !f.first {
		f.first = true
	}

	return err
}

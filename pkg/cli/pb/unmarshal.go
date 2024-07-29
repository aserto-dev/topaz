package pb

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Message[T any] interface {
	proto.Message
	*T
}

func UnmarshalRequest[T any, M Message[T]](src string, msg M) error {
	var bytes []byte

	if src == "-" {
		reader := bufio.NewReader(os.Stdin)
		if b, err := io.ReadAll(reader); err == nil {
			bytes = b
		} else {
			return errors.Wrap(err, "failed to read from stdin")
		}
	} else if fi, err := os.Stat(src); err == nil && !fi.IsDir() {
		if b, err := os.ReadFile(src); err == nil {
			bytes = b
		} else {
			return errors.Wrapf(err, "opening file [%s]", src)
		}
	} else {
		bytes = []byte(src)
	}

	return protojson.Unmarshal(bytes, msg)
}

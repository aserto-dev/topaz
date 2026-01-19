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
	if src == "-" {
		reader := bufio.NewReader(os.Stdin)

		bytes, err := io.ReadAll(reader)
		if err != nil {
			return errors.Wrap(err, "failed to read from stdin")
		}

		return protojson.Unmarshal(bytes, msg)
	}

	if fi, err := os.Stat(src); err == nil && !fi.IsDir() {
		bytes, err := os.ReadFile(src)
		if err != nil {
			return errors.Wrapf(err, "opening file [%s]", src)
		}

		return protojson.Unmarshal(bytes, msg)
	}

	return protojson.Unmarshal([]byte(src), msg)
}

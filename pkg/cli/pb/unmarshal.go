package pb

import (
	"bufio"
	"io"
	"os"

	"github.com/aserto-dev/topaz/pkg/fs"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Message[T any] interface {
	proto.Message
	*T
}

func UnmarshalRequest[T any, M Message[T]](src string, msg M) error {
	switch {
	case src == "":
		return status.Error(codes.InvalidArgument, "empty request")

	case src == "-":
		reader := bufio.NewReader(os.Stdin)
		buf, err := io.ReadAll(reader)
		if err != nil {
			return errors.Wrap(err, "failed to read from stdin")
		}
		return protojson.Unmarshal(buf, msg)

	case fs.FileExists(src):
		buf, err := os.ReadFile(src)
		if err != nil {
			return errors.Wrapf(err, "opening file %q", src)
		}
		return protojson.Unmarshal(buf, msg)

	default:
		return protojson.Unmarshal([]byte(src), msg)
	}
}

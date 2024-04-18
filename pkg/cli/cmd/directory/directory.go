package directory

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type FormatVersion int

const (
	V2 FormatVersion = 2
	V3 FormatVersion = 3
)

type DirectoryCmd struct {
	Import  ImportCmd  `cmd:"" help:"import directory data"`
	Export  ExportCmd  `cmd:"" help:"export directory data"`
	Backup  BackupCmd  `cmd:"" help:"backup directory data"`
	Restore RestoreCmd `cmd:"" help:"restore directory data"`
}

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
	} else if _, err := os.Stat(src); errors.Is(err, os.ErrNotExist) {
		bytes = []byte(src)
	} else {
		if b, err := os.ReadFile(src); err == nil {
			bytes = b
		} else {
			return errors.Wrapf(err, "opening file [%s]", src)
		}
	}

	return protojson.Unmarshal(bytes, msg)
}

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
	Import         ImportCmd         `cmd:"" help:"import data into directory"`
	Export         ExportCmd         `cmd:"" help:"export directory data"`
	GetObject      GetObjectCmd      `cmd:"" help:"get object" group:"directory"`
	SetObject      SetObjectCmd      `cmd:"" help:"set object" group:"directory"`
	DeleteObject   DeleteObjectCmd   `cmd:"" help:"delete object" group:"directory"`
	ListObjects    ListObjectsCmd    `cmd:"" help:"list objects" group:"directory"`
	GetRelation    GetRelationCmd    `cmd:"" help:"get relation" group:"directory"`
	SetRelation    SetRelationCmd    `cmd:"" help:"set relation" group:"directory"`
	DeleteRelation DeleteRelationCmd `cmd:"" help:"delete relation" group:"directory"`
	ListRelations  ListRelationsCmd  `cmd:"" help:"list relations" group:"directory"`
	Check          CheckCmd          `cmd:"" help:"check" group:"directory"`
	GetGraph       GetGraphCmd       `cmd:"" help:"get relation graph" group:"directory"`
}

type Message[T any] interface {
	proto.Message
	*T
}

func UnmarshalRequest[T any, M Message[T]](src string, msg M) error {
	if src == "-" {
		reader := bufio.NewReader(os.Stdin)
		var bytes []byte
		for {
			data, err := reader.ReadByte()
			if err != nil {
				// Read until EOF provided
				if err == io.EOF {
					break
				}
				return err
			}
			bytes = append(bytes, data)
		}

		err := protojson.Unmarshal(bytes, msg)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal request from stdin")
		}
	} else {
		err := protojson.Unmarshal([]byte(src), msg)
		if err != nil {
			dat, err := os.ReadFile(src)
			if err != nil {
				return errors.Wrapf(err, "opening file [%s]", src)
			}

			err = protojson.Unmarshal(dat, msg)
			if err != nil {
				return errors.Wrapf(err, "failed to unmarshal request from file [%s]", src)
			}
		}
	}
	return nil
}

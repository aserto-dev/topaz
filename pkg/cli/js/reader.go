package js

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type Reader struct {
	dec     *json.Decoder
	first   bool
	rootKey string
}

func NewReader(r io.Reader) (*Reader, error) {
	dec := json.NewDecoder(r)

	// advance reader to start token
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	keyStr := ""
	if del, ok := tok.(json.Delim); ok {
		// get key value if not array
		if del == '{' {
			t, err := dec.Token()
			if err != nil {
				return nil, err
			}

			if key, ok := t.(string); ok {
				keyStr = key
			}

			tok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			if delim, ok := tok.(json.Delim); !ok && delim.String() != "[" {
				return nil, errors.Errorf("file does not contain a JSON array")
			}

			return &Reader{
				dec:     dec,
				first:   false,
				rootKey: keyStr,
			}, nil
		}
	}

	return nil, errors.Errorf("unsupported file format")
}

func (r *Reader) GetObjectType() string {
	return r.rootKey
}

func (r *Reader) Close() error {
	return nil
}

func (r *Reader) Read(m proto.Message) error {
	if !r.dec.More() {
		// if no more data in array read ] character at end of array
		tok, err := r.dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := tok.(json.Delim); !ok && delim.String() != "]" {
			return errors.Errorf("file does not contain a JSON array")
		}

		// check json file ends in }
		tok, err = r.dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := tok.(json.Delim); !ok && delim.String() != "}" {
			fmt.Fprintf(os.Stderr, "detected addition data [%s] in file, ignoring.", tok)
		}
		return io.EOF
	}

	if err := pb.UnmarshalNext(r.dec, m); err != nil {
		return err
	}
	return nil
}

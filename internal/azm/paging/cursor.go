package paging

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"

	"github.com/aserto-dev/go-directory/pkg/derr"
)

const (
	// DefaultPageSize default pagination page size.
	DefaultPageSize = int32(100)
	// MinPageSize minimum pagination page size.
	MinPageSize = int32(1)
	// MaxPageSize maximum pagination page size.
	MaxPageSize = int32(100)
	// ServerSetPageSize .
	ServerSetPageSize = int32(0)
	// SingleResultSet (INTERNAL USE ONLY).
	SingleResultSet = int32(-1)
	// TotalsOnlyResultSet .
	TotalsOnlyResultSet = int32(-2)
)

// PageSize validator.
func PageSize(input int32) int32 {
	switch {
	case input == SingleResultSet:
		return SingleResultSet
	case input == TotalsOnlyResultSet:
		return TotalsOnlyResultSet
	case input == ServerSetPageSize:
		return DefaultPageSize
	case input < MinPageSize:
		return MinPageSize
	case input > MaxPageSize:
		return MaxPageSize
	default:
		return input
	}
}

type Cursor struct {
	OptsHash uint64
	Keys     []string
}

func DecodeCursor(encoded string) (*Cursor, error) {
	bin, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, derr.ErrInvalidCursor
	}

	reader := bytes.NewReader(bin)
	dec := gob.NewDecoder(reader)

	var cursor Cursor

	if err := dec.Decode(&cursor); err != nil {
		return nil, derr.ErrInvalidCursor
	}

	return &cursor, nil
}

func (c *Cursor) Encode() (string, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(c); err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(buf.Bytes()), nil
}

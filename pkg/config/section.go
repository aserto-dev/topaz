// Package config contains common types and functions used in defining configuration sections.
package config

import (
	"io"
	"iter"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var ErrConfig = errors.New("configuraion error")

// Serializer can emit a YAML representation of itself into an io.Writer.
type Serializer interface {
	// Serialize as YAML into w.
	Serialize(w io.Writer) error
}

type AccessMode bool

const (
	ReadOnly  AccessMode = false // Read-only access mode.
	ReadWrite AccessMode = true  // Read-write access mode.
)

func (m AccessMode) String() string {
	return lo.Ternary(m == ReadOnly, "ro", "rw")
}

// Section is a configuration element.
type Section interface {
	Serializer

	// Defaults returns the section's default values.
	Defaults() map[string]any

	// Validate determines if the section's values are valid.
	Validate() error

	// Paths retruns a sequence of file system paths referenced in this section.
	// Each element in the sequence is (string, ReadOnly|ReadWrite) pair.
	// The CLI mounts these paths when starting the topaz container.
	//
	// Note: The sequence may be empty but must not be nil. The sequence may contain duplicate values
	// and/or empty strings.
	Paths() iter.Seq2[string, AccessMode]
}

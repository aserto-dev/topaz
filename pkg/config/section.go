package config

import (
	"io"

	"github.com/pkg/errors"
)

var ErrConfig = errors.New("configuraion error")

type Serializer interface {
	// Serialize as YAML into w.
	Serialize(w io.Writer) error
}

// Section is a configuration element.
type Section interface {
	Serializer

	// Defaults returns the section's default values.
	Defaults() map[string]any

	// Validate determines if the section's values are valid.
	Validate() error
}

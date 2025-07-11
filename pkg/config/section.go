// Package config contains common types and functions used in defining configuration sections.
package config

import (
	"io"

	"github.com/pkg/errors"
)

var ErrConfig = errors.New("configuraion error")

// Serializer can emit a YAML representation of itself into an io.Writer.
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

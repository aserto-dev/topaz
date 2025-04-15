package handler

import (
	"io"
)

type Config interface {
	Defaults() map[string]any
	Validate() (bool, error)
	Generate(w io.Writer) error
}

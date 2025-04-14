package handler

import (
	"io"

	"github.com/spf13/viper"
)

type Config interface {
	SetDefaults(v *viper.Viper, p ...string)
	Validate() (bool, error)
	Generate(w io.Writer) error
}

package topaz

import (
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
)

const DefaultRequestTimeout = 5 * time.Second

type Config struct {
	Enabled        bool            `json:"enabled"`
	Connection     aserto.Config   `json:"connection"`
	RequestTimeout config.Duration `json:"request_timeout"`
}

package topaz

import (
	"time"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
)

const DefaultRequestTimeout = 5 * time.Second

type Config struct {
	Enabled        bool            `json:"enabled"`
	Connection     client.Config   `json:"connection"`
	RequestTimeout config.Duration `json:"request_timeout"`
}

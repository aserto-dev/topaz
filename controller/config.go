package controller

import (
	client "github.com/aserto-dev/go-aserto"
)

type Config struct {
	Enabled bool          `json:"enabled"`
	Server  client.Config `json:"server"`
}

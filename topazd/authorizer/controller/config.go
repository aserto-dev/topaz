package controller

import (
	client "github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/pkg/config"
)

type Config struct {
	config.Optional

	Server client.Config `json:"server"`
}

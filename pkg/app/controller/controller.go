package controller

import (
	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/go-aserto/client"

	"github.com/rs/zerolog"
)

func NewControllerFactory(logger *zerolog.Logger, cfg *controller.Config, dop client.DialOptionsProvider) *controller.Factory {
	return controller.NewFactory(logger, cfg, dop)
}

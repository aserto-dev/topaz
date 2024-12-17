package config

import (
	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/debug"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/metrics"
	"github.com/aserto-dev/topaz/pkg/services"
)

const Version int = 3

type Config struct {
	Version        int                   `json:"version"`
	Logging        logger.Config         `json:"logging"`
	Authentication authentication.Config `json:"authentication,omitempty"`
	Debug          debug.Config          `json:"debug,omitempty"`
	Health         health.Config         `json:"health,omitempty"`
	Metrics        metrics.Config        `json:"metrics,omitempty"`
	Services       services.Config       `json:"services"`
	Authorizer     authorizer.Config     `json:"authorizer"`
	Directory      directory.Config      `json:"directory"`
}

type ConfigV3 struct {
	Version int           `json:"version"`
	Logging logger.Config `json:"logging"`
}

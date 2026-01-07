// Package config provides tools for the topaz CLI to consume and generate topazd configuration files.
package config

import (
	"iter"

	cfgutil "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/topazd/loiter"
	"github.com/aserto-dev/topaz/topazd/servers"
)

type Container struct {
	*config.Config
}

// Load reads and validates the configuration file at the given path.
func Load(cfgPath string) (*Container, error) {
	cfg, err := config.Load(cfgPath, config.WithNoEnvSubstitution)
	if err != nil {
		return nil, err
	}

	return &Container{cfg}, nil
}

// Ports returns a sequence of all ports that the loaded configuration listens on.
func (c *Container) Ports() iter.Seq[string] {
	return loiter.Chain(
		optionalPorts(&c.Debug, &c.Health, &c.Metrics),
		serverPorts(c.Servers),
	)
}

func optionalPorts(listeners ...cfgutil.OptionalListener) iter.Seq[string] {
	return func(yield loiter.Yields[string]) {
		for _, l := range listeners {
			if l.IsEnabled() && !yield(l.Port()) {
				break
			}
		}
	}
}

func serverPorts(srvs servers.Config) iter.Seq[string] {
	return func(yield loiter.Yields[string]) {
		for _, s := range srvs {
			if !s.GRPC.IsEmptyAddress() && !yield(s.GRPC.Port()) {
				break
			}

			if !s.HTTP.IsEmptyAddress() && !yield(s.HTTP.Port()) {
				break
			}
		}
	}
}

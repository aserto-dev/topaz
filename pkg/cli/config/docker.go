// Package config provides tools for the topaz CLI to consume and generate topazd configuration files.
package config

import (
	"iter"
	"maps"
	"os"

	"github.com/pkg/errors"

	cfgutil "github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/pkg/loiter"
	"github.com/aserto-dev/topaz/pkg/servers"
)

type containerConfig struct {
	*config.Config
}

// Load reads and validates the configuration file at the given path.
func Load(cfgPath string) (*containerConfig, error) {
	cfg, err := load(cfgPath)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid configuration in %q", cfgPath)
	}

	return &containerConfig{cfg}, nil
}

// Ports returns a sequence of all ports that the loaded configuration listens on.
func (c *containerConfig) Ports() iter.Seq[string] {
	return loiter.Chain(
		optionalPorts(&c.Debug, &c.Health, &c.Metrics),
		serverPorts(maps.Values(c.Servers)),
	)
}

func (c *containerConfig) Volumes() []string {
	return []string{}
}

func load(cfgPath string) (*config.Config, error) {
	f, err := os.Open(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config file %q", cfgPath)
	}

	defer f.Close()

	return config.NewConfig(f)
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

func serverPorts(servers iter.Seq[*servers.Server]) iter.Seq[string] {
	return func(yield loiter.Yields[string]) {
		for s := range servers {
			if !s.GRPC.IsEmptyAddress() && !yield(s.GRPC.Port()) {
				break
			}

			if !s.HTTP.IsEmptyAddress() && !yield(s.HTTP.Port()) {
				break
			}
		}
	}
}

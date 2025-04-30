package topaz

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/aserto-dev/logger"

	"github.com/aserto-dev/topaz/pkg/cli/x"
)

type Server interface {
	Run(ctx context.Context) error
}

type Topaz struct {
	servers []Server
}

func NewTopaz(ctx context.Context, configPath string, configOverrides ...ConfigOverride) (*Topaz, error) {
	cfgBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read config file %q", configPath)
	}

	cfg, err := NewConfig(bytes.NewReader(cfgBytes), configOverrides...)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log, err := logger.NewLogger(os.Stdout, os.Stderr, &cfg.Logging)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cfgBytes), x.EnvTopazDir) {
		log.Warn().Msg("This configuration file uses the obsolete TOPAZ_DIR environment variable.")
		log.Warn().Msg("Please update to use the new TOPAZ_DB_DIR and TOPAZ_CERTS_DIR environment variables.")
	}

	servers, err := newServers(log.WithContext(ctx), cfg)
	if err != nil {
		return nil, err
	}

	return &Topaz{
		servers: servers,
	}, nil
}

func (t *Topaz) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)

	for _, server := range t.servers {
		errGroup.Go(func() error { return server.Run(ctx) })
	}

	return errGroup.Wait()
}

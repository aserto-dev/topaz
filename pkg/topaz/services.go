package topaz

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	client "github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/health"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/service"
	"github.com/aserto-dev/topaz/plugins/edge"
)

type topazServices struct {
	directory  *directory.Service
	authorizer *authorizer.Service
	health     *health.Service
}

func (s *topazServices) Get(ctx context.Context, name servers.ServiceName) service.Service {
	switch name {
	case servers.Service.Access:
		return s.directory.Access()
	case servers.Service.Reader:
		return s.directory.Reader()
	case servers.Service.Writer:
		return s.directory.Writer()
	case servers.Service.Model:
		return s.directory.Model()
	case servers.Service.Importer:
		return s.directory.Importer()
	case servers.Service.Exporter:
		return s.directory.Exporter()
	case servers.Service.Authorizer:
		return s.authorizer
	case servers.Service.Console:
		return service.Noop
	default:
		zerolog.Ctx(ctx).Fatal().Msgf("unknown service %s", name)
		return service.Noop
	}
}

func (s *topazServices) Directory() *directory.Service {
	return s.directory
}

func (s *topazServices) Authorizer() *authorizer.Service {
	return s.authorizer
}

func (s *topazServices) Health() *health.Service {
	return s.health
}

func (s *topazServices) Close(ctx context.Context) error {
	var errs error

	if err := s.authorizer.Close(ctx); err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "failed to close authorizer service"))
	}

	return errs
}

func newTopazServices(ctx context.Context, cfg *config.Config) (*topazServices, error) {
	healthSvc := health.New(&cfg.Health)

	dir, err := newDirectory(ctx, cfg)
	if err != nil {
		return nil, err
	}

	authz, err := newAuthorizer(ctx, cfg, healthSvc)
	if err != nil {
		return nil, err
	}

	return &topazServices{
		directory:  dir,
		authorizer: authz,
		health:     healthSvc,
	}, nil
}

func newDirectory(ctx context.Context, cfg *config.Config) (*directory.Service, error) {
	if !cfg.Servers.DirectoryEnabled() {
		return &directory.Service{}, nil
	}

	return directory.New(ctx, &cfg.Directory)
}

func newAuthorizer(ctx context.Context, cfg *config.Config, healthSvc *health.Service) (*authorizer.Service, error) {
	if !cfg.Servers.ServiceEnabled(servers.Service.Authorizer) {
		return &authorizer.Service{}, nil
	}

	dsCfg := dsConfig(cfg)
	edgeFactory := edge.NewPluginFactory(dsCfg, zerolog.Ctx(ctx), healthSvc)

	az, err := authorizer.New(ctx, &cfg.Authorizer, edgeFactory, dsCfg)
	if err != nil {
		return nil, err
	}

	return az, nil
}

func dsConfig(cfg *config.Config) *client.Config {
	if cfg.Directory.IsRemote() {
		return &cfg.Directory.Store.Remote.Config
	}

	readerServer, _ := cfg.Servers.FindService(servers.Service.Reader)

	return &client.Config{
		Address:    readerServer.GRPC.ListenAddress,
		APIKey:     cfg.Authentication.ReaderKey(),
		CACertPath: lo.Ternary(readerServer.GRPC.Certs.HasCA(), readerServer.GRPC.Certs.CA, ""),
		Insecure:   readerServer.GRPC.Certs.HasCert() && !readerServer.GRPC.Certs.HasCA(),
		NoTLS:      !readerServer.GRPC.Certs.HasCert(),
	}
}

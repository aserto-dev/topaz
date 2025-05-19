package topaz

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	client "github.com/aserto-dev/go-aserto"

	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type topazServices struct {
	directory  *directory.Service
	authorizer *authorizer.Service
}

func (s *topazServices) Directory() *directory.Service {
	return s.directory
}

func (s *topazServices) Authorizer() *authorizer.Service {
	return s.authorizer
}

func (s *topazServices) Close(ctx context.Context) error {
	var errs error

	if err := s.authorizer.Close(ctx); err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "failed to close authorizer service"))
	}

	return errs
}

func newTopazServices(ctx context.Context, cfg *config.Config) (*topazServices, error) {
	dir, err := newDirectory(ctx, cfg)
	if err != nil {
		return nil, err
	}

	authz, err := newAuthorizer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &topazServices{
		directory:  dir,
		authorizer: authz,
	}, nil
}

func newDirectory(ctx context.Context, cfg *config.Config) (*directory.Service, error) {
	if !cfg.Servers.DirectoryEnabled() {
		return &directory.Service{}, nil
	}

	return directory.New(ctx, &cfg.Directory)
}

func newAuthorizer(ctx context.Context, cfg *config.Config) (*authorizer.Service, error) {
	if !cfg.Servers.ServiceEnabled(servers.Service.Authorizer) {
		return &authorizer.Service{}, nil
	}

	dsCfg := dsConfig(cfg)

	az, err := authorizer.New(ctx, &cfg.Authorizer, dsCfg)
	if err != nil {
		return nil, err
	}

	return az, nil
}

func dsConfig(cfg *config.Config) *client.Config {
	if cfg.Directory.IsRemote() {
		return (*client.Config)(&cfg.Directory.Store.Remote)
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

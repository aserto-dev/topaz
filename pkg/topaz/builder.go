package topaz

import (
	"context"
	"net/http"
	"strconv"

	gorilla "github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	cerr "github.com/aserto-dev/errors"

	"github.com/aserto-dev/topaz/pkg/app"
	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/authorizer"
	"github.com/aserto-dev/topaz/pkg/directory"
	"github.com/aserto-dev/topaz/pkg/middleware"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/aserto-dev/topaz/pkg/topaz/config"
)

type Runner interface {
	Go(f func() error)
}

type Server interface {
	Start(ctx context.Context, runner Runner) error
	Stop(ctx context.Context) error
}

type server struct {
	grpc *grpc.Server
	http *http.Server
}

var _ Server = (*server)(nil)

func (s *server) Start(ctx context.Context, runner Runner) error {
	if len(s.grpc.GetServiceInfo()) > 0 {
		// TODO: start grpc server
	}

	return nil
}

func (s *server) Stop(ctx context.Context) error {
	return nil
}

type serverBuilder struct {
	cfg      *config.Config
	services *topazServices

	middleware *middlewares
}

type middlewares struct {
	auth    middleware.Server
	logging *middleware.Logging
}

func (m *middlewares) unary() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(m.logging.Unary(), m.auth.Unary())
}

func (m *middlewares) stream() grpc.ServerOption {
	return grpc.ChainStreamInterceptor(m.logging.Stream(), m.auth.Stream())
}

func newServerBuilder(logger *zerolog.Logger, cfg *config.Config, services *topazServices) *serverBuilder {
	return &serverBuilder{
		cfg:      cfg,
		services: services,
		middleware: &middlewares{
			auth:    authentication.New(&cfg.Authentication),
			logging: middleware.NewLogging(logger),
		},
	}
}

//nolint:ireturn  // Factory function.
func (b *serverBuilder) Build(ctx context.Context, cfg *servers.Server) (Server, error) {
	grpcServer, err := b.buildGRPC(cfg)
	if err != nil {
		return nil, err
	}

	if !cfg.HTTP.HasListener() {
		return &server{grpc: grpcServer}, nil
	}

	httpServer, err := b.buildHTTP(&cfg.HTTP)
	if err != nil {
		return nil, err
	}

	if len(grpcServer.GetServiceInfo()) > 0 {
		// wire up grpc-gateway.
		addr := "dns:///" + cfg.GRPC.ListenAddress
		gwMux := b.gatewayMux(cfg.HTTP.AllowedHeaders)

		creds, err := cfg.GRPC.ClientCredentials()
		if err != nil {
			return nil, err
		}

		for _, service := range cfg.Services {
			if err := b.registerGateway(ctx, service, gwMux, addr, creds); err != nil {
				return nil, err
			}
		}

		apiRouter := httpServer.router.PathPrefix("/api").Subrouter()
		apiRouter.Use(middleware.FieldsMask)
		apiRouter.PathPrefix("/").Handler(gwMux)
	}

	return &server{grpc: grpcServer, http: httpServer.Server}, nil
}

func (b *serverBuilder) buildGRPC(cfg *servers.Server) (*grpc.Server, error) {
	if !cfg.GRPC.HasListener() {
		return &grpc.Server{}, nil
	}

	creds, err := cfg.GRPC.Certs.ServerCredentials()
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(grpc.Creds(creds), b.middleware.unary(), b.middleware.stream())

	// TODO: register reflection service. Need to add a config option.

	for _, service := range cfg.Services {
		b.registerService(server, service)
	}

	return server, nil
}

func (b *serverBuilder) registerService(server *grpc.Server, service servers.ServiceName) {
	switch service {
	case servers.Service.Access:
		b.services.directory.RegisterAccessServer(server)
	case servers.Service.Reader:
		b.services.directory.RegisterReaderServer(server)
	case servers.Service.Writer:
		b.services.directory.RegisterWriterServer(server)
	case servers.Service.Authorizer:
		b.services.authorizer.RegisterAuthorizerServer(server)
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}

func (b *serverBuilder) registerGateway(
	ctx context.Context,
	service servers.ServiceName,
	mux *runtime.ServeMux,
	addr string,
	opts ...grpc.DialOption,
) error {
	switch service {
	case servers.Service.Access:
		return b.services.directory.RegisterAccessGateway(ctx, mux, addr, opts...)
	case servers.Service.Reader:
		return b.services.directory.RegisterReaderGateway(ctx, mux, addr, opts...)
	case servers.Service.Writer:
		return b.services.directory.RegisterWriterGateway(ctx, mux, addr, opts...)
	case servers.Service.Authorizer:
		return b.services.authorizer.RegisterAuthorizerGateway(ctx, mux, addr, opts...)
	default:
		panic(errors.Errorf("unknown service %q", service))
	}
}

func (b *serverBuilder) buildHTTP(cfg *servers.HTTPServer) (*httpServer, error) {
	router := gorilla.NewRouter()

	tlsConf, err := cfg.Certs.ServerConfig()
	if err != nil {
		return nil, err
	}

	return &httpServer{
		Server: &http.Server{
			Addr:              cfg.ListenAddress,
			TLSConfig:         tlsConf,
			Handler:           cfg.Cors().Handler(router),
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		router: router,
	}, nil
}

func (b *serverBuilder) gatewayMux(allowedHeaders []string) *runtime.ServeMux {
	headerSet := lo.SliceToMap(allowedHeaders, func(header string) (string, struct{}) {
		return header, struct{}{}
	})

	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(header string) (string, bool) {
			if _, ok := headerSet[header]; ok {
				return header, true
			}

			return runtime.DefaultHeaderMatcher(header)
		}),
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:          "  ",
					AllowPartial:    true,
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
		runtime.WithMarshalerOption(
			"application/json+masked",
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					Indent:        "  ",
					AllowPartial:  true,
					UseProtoNames: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				},
			},
		),
		runtime.WithUnescapingMode(runtime.UnescapingModeAllExceptSlash),
		runtime.WithForwardResponseOption(forwardXHTTPCode),
		runtime.WithErrorHandler(cerr.CustomErrorHandler),
	)
}

func forwardXHTTPCode(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	headers := metautils.NiceMD(md.HeaderMD)

	// set http status code
	if xcode := headers.Get("x-http-code"); xcode != "" {
		code, err := strconv.Atoi(xcode)
		if err != nil {
			return err
		}
		// delete the headers to not expose any grpc-metadata in http response
		headers.Del("x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	}

	return nil
}

type httpServer struct {
	*http.Server

	router *gorilla.Router
}

type topazServices struct {
	directory  *directory.Service
	authorizer *authorizer.Service
	console    *app.ConsoleService
}

func newTopazServices(ctx context.Context, cfg *config.Config) (*topazServices, error) {
	dir, err := directory.New(ctx, &cfg.Directory)
	if err != nil {
		return nil, err
	}

	return &topazServices{
		directory:  dir,
		authorizer: authorizer.New(ctx, &cfg.Authorizer),
		console:    app.NewConsole(),
	}, nil
}

func countTrue(vals ...bool) int {
	return lo.Reduce(vals,
		func(count int, val bool, _ int) int {
			return count + lo.Ternary(val, 1, 0)
		},
		0,
	)
}

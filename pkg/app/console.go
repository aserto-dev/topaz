package app

import (
	"context"
	"fmt"
	"strings"

	builder "github.com/aserto-dev/service-host"
	"github.com/aserto-dev/topaz/pkg/app/handlers"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ConsoleService struct{}

const (
	consoleService = "console"
)

func NewConsole() ServiceTypes {
	return &ConsoleService{}
}

func (e *ConsoleService) AvailableServices() []string {
	return []string{"console"}
}

func (e *ConsoleService) GetGRPCRegistrations(services ...string) builder.GRPCRegistrations {
	return func(server *grpc.Server) {
	}
}

func (e *ConsoleService) GetGatewayRegistration(services ...string) builder.HandlerRegistrations {
	return func(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string, opts []grpc.DialOption) error {
		return nil
	}
}

func (e *ConsoleService) Cleanups() []func() {
	return nil
}

func (e *ConsoleService) PrepareConfig(cfg *config.Config) *handlers.ConsoleCfg {
	authorizerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[authorizerService]; ok {
		authorizerURL = getGatewayAddress(serviceConfig)
	}
	readerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[readerService]; ok {
		readerURL = getGatewayAddress(serviceConfig)
	}
	writerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[writerService]; ok {
		writerURL = getGatewayAddress(serviceConfig)
	}
	modelURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[modelService]; ok {
		modelURL = getGatewayAddress(serviceConfig)
	}

	authorizerAPIKey := ""
	if _, ok := cfg.APIConfig.Services[authorizerService]; ok {
		for key := range cfg.Auth.APIKeys {
			// we only need a key
			authorizerAPIKey = key
			break
		}
	}

	directoryAPIKey := ""
	if _, ok := cfg.APIConfig.Services[readerService]; ok {
		for key := range cfg.Auth.APIKeys {
			// we only need a key
			directoryAPIKey = key
			break
		}
	}

	return &handlers.ConsoleCfg{
		AsertoDirectoryURL:       readerURL,
		DirectoryAPIKey:          directoryAPIKey,
		DirectoryTenantID:        cfg.DirectoryResolver.APIKey,
		AuthorizerServiceURL:     authorizerURL,
		AuthorizerAPIKey:         authorizerAPIKey,
		AsertoDirectoryReaderURL: &readerURL,
		AsertoDirectoryWriterURL: &writerURL,
		AsertoDirectoryModelURL:  &modelURL,
	}
}

func getGatewayAddress(serviceConfig *builder.API) string {
	addr := serviceAddress(serviceConfig.Gateway.ListenAddress)

	if serviceConfig.Gateway.HTTP {
		return fmt.Sprintf("http://%s", addr)
	} else {
		return fmt.Sprintf("https://%s", addr)
	}
}

func serviceAddress(listenAddress string) string {
	return strings.Replace(listenAddress, "0.0.0.0", "localhost", 1)
}

package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aserto-dev/topaz/pkg/app/handlers"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	builder "github.com/aserto-dev/topaz/pkg/service/builder"
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
		return mux.HandlePath("GET", "/", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			http.Redirect(w, r, "/ui/directory/model", http.StatusSeeOther)
		}))
	}
}

func (e *ConsoleService) Cleanups() []func() {
	return nil
}

func (e *ConsoleService) PrepareConfig(cfg *config.Config) *handlers.TopazCfg {
	directoryServiceURL := serviceAddress(fmt.Sprintf("https://%s", strings.Split(cfg.DirectoryResolver.Address, ":")[0]))

	authorizerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[authorizerService]; ok {
		authorizerURL = getGatewayAddress(serviceConfig)
	}

	readerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[readerService]; ok {
		readerURL = getGatewayAddress(serviceConfig)
		if cfg.DirectoryResolver.Address == serviceConfig.GRPC.ListenAddress {
			directoryServiceURL = readerURL
		}
	}

	writerURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[writerService]; ok {
		writerURL = getGatewayAddress(serviceConfig)
	}

	importerURL := "" // always empty, no gateway service associated with the importer service.

	exporterURL := "" // always empty, no gateway service associated with the exporter service.

	modelURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[modelService]; ok {
		modelURL = getGatewayAddress(serviceConfig)
	}

	consoleURL := ""
	if serviceConfig, ok := cfg.APIConfig.Services[consoleService]; ok {
		consoleURL = getGatewayAddress(serviceConfig)
	}

	authorizerAPIKey := ""
	if _, ok := cfg.APIConfig.Services[authorizerService]; ok {
		for key := range cfg.Auth.APIKeys {
			// we only need a key
			authorizerAPIKey = key
			break
		}
	}

	directoryAPIKey := cfg.DirectoryResolver.APIKey

	return &handlers.TopazCfg{
		AuthorizerServiceURL:        authorizerURL,
		AuthorizerAPIKey:            authorizerAPIKey,
		DirectoryServiceURL:         directoryServiceURL,
		DirectoryAPIKey:             directoryAPIKey,
		DirectoryTenantID:           cfg.DirectoryResolver.TenantID,
		DirectoryReaderServiceURL:   readerURL,
		DirectoryWriterServiceURL:   writerURL,
		DirectoryImporterServiceURL: importerURL,
		DirectoryExporterServiceURL: exporterURL,
		DirectoryModelServiceURL:    modelURL,
		ConsoleURL:                  consoleURL,
	}
}

func getGatewayAddress(serviceConfig *builder.API) string {
	if serviceConfig.Gateway.FQDN != "" {
		return serviceConfig.Gateway.FQDN
	}
	addr := serviceAddress(serviceConfig.Gateway.ListenAddress)

	serviceConfig.Gateway.HTTP = !serviceConfig.Gateway.Certs.HasCert()

	if serviceConfig.Gateway.HTTP {
		return fmt.Sprintf("http://%s", addr)
	}
	return fmt.Sprintf("https://%s", addr)
}

func serviceAddress(listenAddress string) string {
	return strings.Replace(listenAddress, "0.0.0.0", "localhost", 1)
}

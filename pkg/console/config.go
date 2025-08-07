package console

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aserto-dev/topaz/pkg/authentication"
	"github.com/aserto-dev/topaz/pkg/config/v3"
	"github.com/aserto-dev/topaz/pkg/servers"
	"github.com/samber/lo"
)

type URLs struct {
	AuthorizerServiceURL      string `json:"authorizerServiceUrl"`
	DirectoryServiceURL       string `json:"directoryServiceUrl"`
	DirectoryReaderServiceURL string `json:"directoryReaderServiceUrl,omitempty"`
	DirectoryWriterServiceURL string `json:"directoryWriterServiceUrl,omitempty"`
	DirectoryModelServiceURL  string `json:"directoryModelServiceUrl,omitempty"`
}

type urlsWithKeys struct {
	*URLs
	AuthorizerAPIKey string `json:"authorizerApiKey"`
	DirectoryAPIKey  string `json:"directoryApiKey"`
}

type configResponse struct {
	ReadOnly           bool            `json:"readOnly"`
	AuthenticationType string          `json:"authenticationType"`
	Configs            []*urlsWithKeys `json:"configs"`
}

func ConfigHandler(topazCfg *config.Config) http.Handler {
	readerURL := gatewayURL(topazCfg.Servers, servers.Service.Reader)

	urls := URLs{
		AuthorizerServiceURL:      gatewayURL(topazCfg.Servers, servers.Service.Authorizer),
		DirectoryServiceURL:       readerURL,
		DirectoryReaderServiceURL: readerURL,
		DirectoryWriterServiceURL: gatewayURL(topazCfg.Servers, servers.Service.Writer),
		DirectoryModelServiceURL:  gatewayURL(topazCfg.Servers, servers.Service.Model),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := apiKey(r)
		cfg := urlsWithKeys{&urls, key, key}
		resp := configResponse{true, authType(&topazCfg.Authentication), []*urlsWithKeys{&cfg}}
		buf, _ := json.Marshal(resp)

		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))

		_, _ = w.Write(buf)
	})
}

func gatewayURL(cfg servers.Config, svc servers.ServiceName) string {
	server, found := cfg.FindService(svc)
	if !found || server.HTTP.IsEmptyAddress() {
		return ""
	}

	if server.HTTP.HostedDomain != "" {
		return server.HTTP.HostedDomain
	}

	scheme := lo.Ternary(server.HTTP.Certs.HasCert(), "https", "http")
	_, port, _ := strings.Cut(server.HTTP.ListenAddress, ":")

	return scheme + "://localhost:" + port
}

func apiKey(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// we know the header is syntactically valid because it got through the authentication middleware.
	_, key, _ := strings.Cut(authHeader, " ")

	return key
}

func authType(cfg *authentication.Config) string {
	if cfg.Enabled {
		return "apiKey"
	}

	return "anonymous"
}

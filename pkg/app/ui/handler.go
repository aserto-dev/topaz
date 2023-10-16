package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aserto-dev/topaz/pkg/cc/config"
)

type fsWithDefinition struct {
	consoleFS http.FileSystem
}

type consoleCfg struct {
	AsertoDirectoryURL       string `json:"asertoDirectoryUrl"`
	AuthorizerServiceURL     string `json:"authorizerServiceUrl"`
	AuthorizerAPIKey         string `json:"authorizerApiKey"`
	DirectoryAPIKey          string `json:"directoryApiKey"`
	DirectoryTenantID        string `json:"directoryTenantId"`
	AsertoDirectoryReaderURL string `json:"asertoDirectoryReaderUrl"`
	AsertoDirectoryWriterURL string `json:"asertoDirectoryWriterUrl"`
	AsertoDirectoryModelURL  string `json:"asertoDirectoryModelUrl"`
}

type consoleCfgWithRemoteDirectory struct {
	AsertoDirectoryURL   string `json:"asertoDirectoryUrl"`
	AuthorizerServiceURL string `json:"authorizerServiceUrl"`
	AuthorizerAPIKey     string `json:"authorizerApiKey"`
	DirectoryAPIKey      string `json:"directoryApiKey"`
	DirectoryTenantID    string `json:"directoryTenantId"`
}

func (f *fsWithDefinition) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/ui/") {
		return f.consoleFS.Open("console/index.html")
	}

	name = strings.TrimPrefix(name, "/public")
	return f.consoleFS.Open(fmt.Sprintf("console%s", name))
}

func UIHandler(consoleFS http.FileSystem) http.Handler {
	return http.FileServer(&fsWithDefinition{consoleFS: consoleFS})
}

func ConfigHandler(confServices *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiKey string
		for key := range confServices.Auth.APIKeys {
			apiKey = key
			break
		}

		if _, ok := confServices.APIConfig.Services["reader"]; ok {
			cfg := composeConfig(confServices, apiKey)
			buf, _ := json.Marshal(cfg)
			writeJSON(buf, w, r)
		} else {
			cfg := composeRemoteDiretoryConfig(confServices, apiKey)
			buf, _ := json.Marshal(cfg)
			writeJSON(buf, w, r)
		}
	}
}

func composeConfig(confServices *config.Config, apiKey string) *consoleCfg {
	cfg := &consoleCfg{}
	cfg.AsertoDirectoryReaderURL = fmt.Sprintf("https://%s", serviceAddress(confServices.APIConfig.Services["reader"].Gateway.ListenAddress))
	cfg.AsertoDirectoryURL = fmt.Sprintf("https://%s", serviceAddress(confServices.APIConfig.Services["reader"].Gateway.ListenAddress))

	if confServices.APIConfig.Services["writer"] != nil {
		cfg.AsertoDirectoryWriterURL = fmt.Sprintf("https://%s", serviceAddress(confServices.APIConfig.Services["writer"].Gateway.ListenAddress))
	}

	if confServices.APIConfig.Services["model"] != nil {
		cfg.AsertoDirectoryModelURL = fmt.Sprintf("https://%s", serviceAddress(confServices.APIConfig.Services["model"].Gateway.ListenAddress))
	}

	if serviceConfig, ok := confServices.APIConfig.Services["authorizer"]; ok {
		cfg.AuthorizerServiceURL = fmt.Sprintf("https://%s", serviceAddress(serviceConfig.Gateway.ListenAddress))
		cfg.AuthorizerAPIKey = apiKey
	}

	if confServices.DirectoryResolver.APIKey != "" {
		cfg.DirectoryAPIKey = confServices.DirectoryResolver.APIKey
	}

	return cfg
}

func serviceAddress(listenAddress string) string {
	addr, port, found := strings.Cut(listenAddress, ":")
	if addr == "0.0.0.0" {
		addr = "localhost"
	}

	if found {
		return fmt.Sprintf("%s:%s", addr, port)
	}

	return addr
}

func composeRemoteDiretoryConfig(confServices *config.Config, apiKey string) *consoleCfgWithRemoteDirectory {
	cfg := &consoleCfgWithRemoteDirectory{}
	host := strings.Split(confServices.DirectoryResolver.Address, ":")[0]
	cfg.AsertoDirectoryURL = fmt.Sprintf("https://%s", host)
	if confServices.DirectoryResolver.TenantID != "" {
		cfg.DirectoryTenantID = confServices.DirectoryResolver.TenantID
	}

	if serviceConfig, ok := confServices.APIConfig.Services["authorizer"]; ok {
		cfg.AuthorizerServiceURL = fmt.Sprintf("https://%s", serviceAddress(serviceConfig.Gateway.ListenAddress))
		cfg.AuthorizerAPIKey = apiKey
	}

	if confServices.DirectoryResolver.APIKey != "" {
		cfg.DirectoryAPIKey = confServices.DirectoryResolver.APIKey
	}

	return cfg
}

func AuthorizersHandler(confServices *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type AuthorizerInstance struct {
			Name   string `json:"name"`
			URL    string `json:"url"`
			APIKey string `json:"apiKey"`
		}
		type authorizersResult struct {
			Results []AuthorizerInstance `json:"results"`
		}

		var apiKey string
		for key := range confServices.Auth.APIKeys {
			apiKey = key
			break
		}

		var cfg *authorizersResult
		if serviceConfig, ok := confServices.APIConfig.Services["authorizer"]; ok {
			cfg = &authorizersResult{
				Results: []AuthorizerInstance{{URL: fmt.Sprintf("https://%s", serviceConfig.Gateway.ListenAddress), Name: "authorizer", APIKey: apiKey}},
			}
		} else {
			cfg = &authorizersResult{}
		}

		buf, _ := json.Marshal(cfg)
		writeJSON(buf, w, r)
	}
}

func writeJSON(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

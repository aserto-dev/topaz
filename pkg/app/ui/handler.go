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

func (f *fsWithDefinition) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/ui/") {
		return f.consoleFS.Open("console/build/index.html")
	}

	name = strings.TrimPrefix(name, "/public")
	return f.consoleFS.Open(fmt.Sprintf("console/build%s", name))
}

func UIHandler(consoleFS http.FileSystem) http.Handler {
	return http.FileServer(&fsWithDefinition{consoleFS: consoleFS})
}

func ConfigHandler(confServices *config.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type consoleCfg struct {
			AsertoDirectoryUrl   string `json:"asertoDirectoryUrl"`
			AuthorizerServiceUrl string `json:"authorizerServiceUrl"`
			AuthorizerApiKey     string `json:"authorizerApiKey"`
			DirectoryApiKey      string `json:"directoryApiKey"`
			DirectoryTenantId    string `json:"directoryTenantId"`
		}

		var apiKey string
		for key := range confServices.Auth.APIKeys {
			apiKey = key
			break
		}

		cfg := &consoleCfg{}
		if serviceConfig, ok := confServices.Services["authorizer"]; ok {
			cfg.AuthorizerServiceUrl = fmt.Sprintf("https://%s", serviceConfig.Gateway.ListenAddress)
			cfg.AuthorizerApiKey = apiKey
		}

		if serviceConfig, ok := confServices.Services["reader"]; ok {
			cfg.AsertoDirectoryUrl = fmt.Sprintf("https://%s", serviceConfig.Gateway.ListenAddress)
		} else {
			host := strings.Split(confServices.DirectoryResolver.Address, ":")[0]
			cfg.AsertoDirectoryUrl = fmt.Sprintf("https://%s", host)
			if confServices.DirectoryResolver.TenantID != "" {
				cfg.DirectoryTenantId = confServices.DirectoryResolver.TenantID
			}
		}

		if confServices.DirectoryResolver.APIKey != "" {
			cfg.DirectoryApiKey = confServices.DirectoryResolver.APIKey
		}

		buf, _ := json.Marshal(cfg)
		writeFile(buf, w, r)
	}
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
		if serviceConfig, ok := confServices.Services["authorizer"]; ok {
			cfg = &authorizersResult{
				Results: []AuthorizerInstance{{URL: fmt.Sprintf("https://%s", serviceConfig.Gateway.ListenAddress), Name: "authorizer", APIKey: apiKey}},
			}
		} else {
			cfg = &authorizersResult{}
		}

		buf, _ := json.Marshal(cfg)
		writeFile(buf, w, r)
	}
}

func writeFile(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

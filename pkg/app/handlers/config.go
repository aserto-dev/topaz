package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aserto-dev/header"
)

var (
	AuthEnabled       = header.CtxKey("AuthEnabled")
	AuthenticatedUser = header.CtxKey("AuthenticatedUser")
)

type ConsoleCfg struct {
	AuthorizerServiceURL        string  `json:"authorizerServiceUrl"`
	AuthorizerAPIKey            string  `json:"authorizerApiKey"`
	DirectoryServiceURL         string  `json:"directoryServiceUrl"`
	DirectoryAPIKey             string  `json:"directoryApiKey"`
	DirectoryTenantID           string  `json:"directoryTenantId"`
	DirectoryReaderServiceURL   *string `json:"directoryReaderServiceUrl,omitempty"`
	DirectoryWriterServiceURL   *string `json:"directoryWriterServiceUrl,omitempty"`
	DirectoryImporterServiceURL *string `json:"directoryImporterServiceUrl,omitempty"`
	DirectoryExporterServiceURL *string `json:"directoryExporterServiceUrl,omitempty"`
	DirectoryModelServiceURL    *string `json:"directoryModelServiceUrl,omitempty"`
}

type ConsoleCfgV1 struct {
	AsertoDirectoryURL       string  `json:"asertoDirectoryUrl"`
	AuthorizerServiceURL     string  `json:"authorizerServiceUrl"`
	AuthorizerAPIKey         string  `json:"authorizerApiKey"`
	DirectoryAPIKey          string  `json:"directoryApiKey"`
	DirectoryTenantID        string  `json:"directoryTenantId"`
	AsertoDirectoryReaderURL *string `json:"asertoDirectoryReaderUrl,omitempty"`
	AsertoDirectoryWriterURL *string `json:"asertoDirectoryWriterUrl,omitempty"`
	AsertoDirectoryModelURL  *string `json:"asertoDirectoryModelUrl,omitempty"`
}

type ConsoleCfgV2 struct {
	Type    string `json:"configType"`
	Name    string `json:"name"`
	Address string `json:"address"`
	*ConsoleCfg
}

type CfgV2Response struct {
	ReadOnly              bool            `json:"readOnly"`
	AuthenticationEnabled bool            `json:"authenticationEnabled"`
	Configs               []*ConsoleCfgV2 `json:"configs"`
}

func ConfigHandler(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		v1Cfg := &ConsoleCfgV1{
			AsertoDirectoryURL:       confServices.DirectoryServiceURL,
			AuthorizerServiceURL:     confServices.AuthorizerServiceURL,
			AuthorizerAPIKey:         confServices.AuthorizerAPIKey,
			DirectoryAPIKey:          confServices.DirectoryAPIKey,
			DirectoryTenantID:        confServices.DirectoryTenantID,
			AsertoDirectoryReaderURL: confServices.DirectoryReaderServiceURL,
			AsertoDirectoryWriterURL: confServices.DirectoryWriterServiceURL,
			AsertoDirectoryModelURL:  confServices.DirectoryModelServiceURL,
		}

		buf, _ := json.Marshal(v1Cfg)
		writeJSON(buf, w, r)
	}
}

func ConfigHandlerV2(confServices *ConsoleCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizerAPIKey := ""
		directoryAPIKey := ""
		authenticatedUser := r.Context().Value(AuthenticatedUser)
		if authenticatedUser != nil && authenticatedUser.(bool) {
			authorizerAPIKey = confServices.AuthorizerAPIKey
			directoryAPIKey = confServices.DirectoryAPIKey
		}

		cfg := &ConsoleCfg{
			AuthorizerServiceURL:        confServices.AuthorizerServiceURL,
			AuthorizerAPIKey:            authorizerAPIKey,
			DirectoryServiceURL:         confServices.DirectoryServiceURL,
			DirectoryAPIKey:             directoryAPIKey,
			DirectoryTenantID:           confServices.DirectoryTenantID,
			DirectoryReaderServiceURL:   confServices.DirectoryReaderServiceURL,
			DirectoryWriterServiceURL:   confServices.DirectoryWriterServiceURL,
			DirectoryImporterServiceURL: confServices.DirectoryImporterServiceURL,
			DirectoryExporterServiceURL: confServices.DirectoryExporterServiceURL,
			DirectoryModelServiceURL:    confServices.DirectoryModelServiceURL,
		}

		cfgV2 := &ConsoleCfgV2{
			Type:       "auto",
			Name:       "Topaz Config",
			Address:    "https://localhost:4321/api/v2/config",
			ConsoleCfg: cfg,
		}

		cfgV2Response := &CfgV2Response{
			Configs:  []*ConsoleCfgV2{cfgV2},
			ReadOnly: true,
		}

		authEnabled := r.Context().Value(AuthEnabled)
		if authEnabled != nil {
			cfgV2Response.AuthenticationEnabled = authEnabled.(bool)
		} else {
			cfgV2Response.AuthenticationEnabled = false
		}

		buf, _ := json.Marshal(cfgV2Response)
		writeJSON(buf, w, r)
	})
}

func writeJSON(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

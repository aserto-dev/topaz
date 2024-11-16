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

type TopazCfg struct {
	AuthorizerServiceURL        string `json:"authorizerServiceUrl"`
	AuthorizerAPIKey            string `json:"authorizerApiKey"`
	DirectoryServiceURL         string `json:"directoryServiceUrl"`
	DirectoryAPIKey             string `json:"directoryApiKey"`
	DirectoryTenantID           string `json:"directoryTenantId"`
	DirectoryReaderServiceURL   string `json:"directoryReaderServiceUrl,omitempty"`
	DirectoryWriterServiceURL   string `json:"directoryWriterServiceUrl,omitempty"`
	DirectoryImporterServiceURL string `json:"directoryImporterServiceUrl,omitempty"`
	DirectoryExporterServiceURL string `json:"directoryExporterServiceUrl,omitempty"`
	DirectoryModelServiceURL    string `json:"directoryModelServiceUrl,omitempty"`
	ConsoleURL                  string `json:"-"`
}

type TopazCfgV1 struct {
	AsertoDirectoryURL       string `json:"asertoDirectoryUrl"`
	AuthorizerServiceURL     string `json:"authorizerServiceUrl"`
	AuthorizerAPIKey         string `json:"authorizerApiKey"`
	DirectoryAPIKey          string `json:"directoryApiKey"`
	DirectoryTenantID        string `json:"directoryTenantId"`
	AsertoDirectoryReaderURL string `json:"asertoDirectoryReaderUrl,omitempty"`
	AsertoDirectoryWriterURL string `json:"asertoDirectoryWriterUrl,omitempty"`
	AsertoDirectoryModelURL  string `json:"asertoDirectoryModelUrl,omitempty"`
}

type TopazCfgV2 struct {
	Type    string `json:"configType"`
	Name    string `json:"name"`
	Address string `json:"address"`
	*TopazCfg
}

type CfgV2Response struct {
	ReadOnly           bool          `json:"readOnly"`
	AuthenticationType string        `json:"authenticationType"`
	Configs            []*TopazCfgV2 `json:"configs"`
}

func ConfigHandler(confServices *TopazCfg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		v1Cfg := &TopazCfgV1{
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

func ConfigHandlerV2(confServices *TopazCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizerAPIKey := ""
		directoryAPIKey := ""
		authenticatedUser := r.Context().Value(AuthenticatedUser)
		if authenticatedUser != nil && authenticatedUser.(bool) {
			authorizerAPIKey = confServices.AuthorizerAPIKey
			directoryAPIKey = confServices.DirectoryAPIKey
		}

		cfg := &TopazCfg{
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

		cfgV2 := &TopazCfgV2{
			Type:     "auto",
			Name:     "Topaz Config",
			Address:  confServices.ConsoleURL + "/api/v2/config",
			TopazCfg: cfg,
		}

		cfgV2Response := &CfgV2Response{
			Configs:            []*TopazCfgV2{cfgV2},
			ReadOnly:           true,
			AuthenticationType: authType(r),
		}

		buf, _ := json.Marshal(cfgV2Response)
		writeJSON(buf, w, r)
	})
}

func authType(r *http.Request) string {
	authEnabled := r.Context().Value(AuthEnabled)
	if authEnabled != nil && authEnabled.(bool) {
		return "apiKey"
	}
	return "anonymous"
}

func writeJSON(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

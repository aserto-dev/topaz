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
	Type                      string  `json:"configType"`
	Name                      string  `json:"name"`
	Address                   string  `json:"address"`
	AuthorizerServiceURL      string  `json:"authorizerServiceUrl"`
	AuthorizerAPIKey          string  `json:"authorizerApiKey"`
	DirectoryServiceURL       *string `json:"directoryServiceUrl,omitempty"`
	DirectoryAPIKey           string  `json:"directoryApiKey"`
	DirectoryTenantID         string  `json:"directoryTenantId"`
	DirectoryReaderServiceURL *string `json:"directoryReaderServiceUrl,omitempty"`
	DirectoryWriterServiceURL *string `json:"directoryWriterServiceUrl,omitempty"`
	DirectoryModelServiceURL  *string `json:"directoryModelServiceUrl,omitempty"`
}

type CfgV2Response struct {
	ReadOnly              bool            `json:"readOnly"`
	AuthenticationEnabled bool            `json:"authenticationEnabled"`
	AuthenticatedUser     bool            `json:"authenticatedUser"`
	Configs               []*ConsoleCfgV2 `json:"configs"`
}

func ConfigHandler(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := json.Marshal(confServices)
		writeJSON(buf, w, r)
	}
}

func ConfigHandlerV2(confServices *ConsoleCfg) http.Handler {
	cfgV2 := &ConsoleCfgV2{
		Type:                      "auto",
		Name:                      "Topaz Config",
		Address:                   "https://localhost:4321/api/v2/config",
		AuthorizerServiceURL:      confServices.AuthorizerServiceURL,
		AuthorizerAPIKey:          "",
		DirectoryAPIKey:           "",
		DirectoryTenantID:         confServices.DirectoryTenantID,
		DirectoryReaderServiceURL: confServices.AsertoDirectoryReaderURL,
		DirectoryWriterServiceURL: confServices.AsertoDirectoryWriterURL,
		DirectoryModelServiceURL:  confServices.AsertoDirectoryModelURL,
	}

	cfgV2Response := &CfgV2Response{
		Configs:  []*ConsoleCfgV2{cfgV2},
		ReadOnly: true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authEnabled := r.Context().Value(AuthEnabled)
		if authEnabled != nil {
			cfgV2Response.AuthenticationEnabled = authEnabled.(bool)
		} else {
			cfgV2Response.AuthenticationEnabled = false
		}

		authenticatedUser := r.Context().Value(AuthenticatedUser)
		if authenticatedUser != nil && authenticatedUser.(bool) {
			cfgV2Response.AuthenticatedUser = true
			cfgV2Response.Configs[0].AuthorizerAPIKey = confServices.AuthorizerAPIKey
			cfgV2Response.Configs[0].DirectoryAPIKey = confServices.DirectoryAPIKey
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

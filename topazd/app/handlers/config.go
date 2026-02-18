package handlers

import (
	"context"
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
	DirectoryReaderServiceURL   string `json:"directoryReaderServiceUrl,omitempty"`
	DirectoryWriterServiceURL   string `json:"directoryWriterServiceUrl,omitempty"`
	DirectoryImporterServiceURL string `json:"directoryImporterServiceUrl,omitempty"`
	DirectoryExporterServiceURL string `json:"directoryExporterServiceUrl,omitempty"`
	DirectoryModelServiceURL    string `json:"directoryModelServiceUrl,omitempty"`
	ConsoleURL                  string `json:"-"`
}

type TopazCfgV2 struct {
	*TopazCfg

	Type    string `json:"configType"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type CfgV2Response struct {
	ReadOnly           bool          `json:"readOnly"`
	AuthenticationType string        `json:"authenticationType"`
	Configs            []*TopazCfgV2 `json:"configs"`
}

func ConfigHandlerV2(confServices *TopazCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizerAPIKey := ""
		directoryAPIKey := ""

		if authenticatedUser, ok := getValueFromCtx[bool](r.Context(), AuthenticatedUser); ok && authenticatedUser {
			authorizerAPIKey = confServices.AuthorizerAPIKey
			directoryAPIKey = confServices.DirectoryAPIKey
		}

		cfg := &TopazCfg{
			AuthorizerServiceURL:        confServices.AuthorizerServiceURL,
			AuthorizerAPIKey:            authorizerAPIKey,
			DirectoryServiceURL:         confServices.DirectoryServiceURL,
			DirectoryAPIKey:             directoryAPIKey,
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

const (
	AuthTypeAnonymous string = "anonymous"
	AuthTypeAPIKey    string = "apiKey"
)

func authType(r *http.Request) string {
	if authEnabled, ok := getValueFromCtx[bool](r.Context(), AuthEnabled); !ok || !authEnabled {
		return AuthTypeAnonymous
	}

	return AuthTypeAPIKey
}

func writeJSON(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

func getValueFromCtx[T any](ctx context.Context, k header.CtxKey) (T, bool) {
	val := ctx.Value(k)
	if val == nil {
		var zero T
		return zero, false
	}

	typedVal, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return typedVal, ok
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	Type    string `json:"configType"`
	Name    string `json:"name"`
	Address string `json:"address"`
	*ConsoleCfg
}

type CfgV2Response struct {
	ReadOnly bool            `json:"readOnly"`
	Configs  []*ConsoleCfgV2 `json:"configs"`
}

func ConfigHandler(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := json.Marshal(confServices)
		writeJSON(buf, w, r)
	}
}

func ConfigHandlerV2(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	cfgV2 := &ConsoleCfgV2{
		Type:       "auto",
		Name:       "Topaz Config",
		Address:    "https://localhost:4321/api/v2/config",
		ConsoleCfg: confServices,
	}

	cfgV2Response := &CfgV2Response{
		Configs:  []*ConsoleCfgV2{cfgV2},
		ReadOnly: true,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := json.Marshal(cfgV2Response)
		writeJSON(buf, w, r)
	}
}

func writeJSON(buf []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Content-Length", strconv.FormatInt(int64(len(buf)), 10))
	_, _ = w.Write(buf)
}

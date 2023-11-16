package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type fsWithDefinition struct {
	consoleFS http.FileSystem
}

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

func ConfigHandler(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	confServices.AuthorizerServiceURL = serviceAddress(confServices.AuthorizerServiceURL)
	confServices.AsertoDirectoryURL = serviceAddress(confServices.AsertoDirectoryURL)

	asertoDirectoryModelURL := serviceAddress(*confServices.AsertoDirectoryModelURL)
	confServices.AsertoDirectoryModelURL = &asertoDirectoryModelURL

	asertoDirectoryReaderURL := serviceAddress(*confServices.AsertoDirectoryReaderURL)
	confServices.AsertoDirectoryReaderURL = &asertoDirectoryReaderURL

	asertoDirectoryWriterURL := serviceAddress(*confServices.AsertoDirectoryWriterURL)
	confServices.AsertoDirectoryWriterURL = &asertoDirectoryWriterURL

	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := json.Marshal(confServices)
		writeJSON(buf, w, r)
	}
}

func serviceAddress(listenAddress string) string {
	return strings.Replace(listenAddress, "0.0.0.0", "localhost", 1)
}

func AuthorizersHandler(confServices *ConsoleCfg) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type AuthorizerInstance struct {
			Name   string `json:"name"`
			URL    string `json:"url"`
			APIKey string `json:"apiKey"`
		}
		type authorizersResult struct {
			Results []AuthorizerInstance `json:"results"`
		}

		cfg := &authorizersResult{
			Results: []AuthorizerInstance{{URL: confServices.AuthorizerServiceURL, Name: "authorizer", APIKey: confServices.AuthorizerAPIKey}},
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

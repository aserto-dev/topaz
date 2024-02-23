package handlers

import (
	"encoding/json"
	"net/http"
)

func AuthorizersHandler(confServices *TopazCfg) func(w http.ResponseWriter, r *http.Request) {
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

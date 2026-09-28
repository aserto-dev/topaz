package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const AuthZENConfiguration string = `/.well-known/authzen-configuration`

type WellKnownConfig struct {
	PolicyDecisionPoint       string `json:"policy_decision_point"`       //nolint:tagliatelle
	AccessEvaluationEndpoint  string `json:"access_evaluation_endpoint"`  //nolint:tagliatelle
	AccessEvaluationsEndpoint string `json:"access_evaluations_endpoint"` //nolint:tagliatelle
	SearchSubjectEndpoint     string `json:"search_subject_endpoint"`     //nolint:tagliatelle
	SearchResourceEndpoint    string `json:"search_resource_endpoint"`    //nolint:tagliatelle
	SearchActionEndpoint      string `json:"search_action_endpoint"`      //nolint:tagliatelle
}

func WellKnownConfigHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		endpoint := &url.URL{
			Scheme: scheme(r),
			Host:   r.Host,
		}

		config := WellKnownConfig{
			PolicyDecisionPoint:       endpoint.String(),
			AccessEvaluationEndpoint:  endpoint.String() + "/access/v1/evaluation",
			AccessEvaluationsEndpoint: endpoint.String() + "/access/v1/evaluations",
			SearchSubjectEndpoint:     endpoint.String() + "/access/v1/search/subject",
			SearchResourceEndpoint:    endpoint.String() + "/access/v1/search/resource",
			SearchActionEndpoint:      endpoint.String() + "/access/v1/search/action",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	}
}

func WellKnownConfigRuntimeHandler() func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	handler := WellKnownConfigHandler()

	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		handler(w, r)
	}
}

func scheme(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}

	return scheme
}

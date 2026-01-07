package directory

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const AuthZENConfiguration string = `/.well-known/authzen-configuration`

type WellKnownConfig struct {
	PolicyDecisionPoint       string `json:"policy_decision_point"`
	AccessEvaluationEndpoint  string `json:"access_evaluation_endpoint"`
	AccessEvaluationsEndpoint string `json:"access_evaluations_endpoint"`
	SearchSubjectEndpoint     string `json:"search_subject_endpoint"`
	SearchResourceEndpoint    string `json:"search_resource_endpoint"`
	SearchActionEndpoint      string `json:"search_action_endpoint"`
}

func WellKnownConfigHandler(endpoint *url.URL) http.HandlerFunc {
	baseURL := endpoint.String()

	return func(w http.ResponseWriter, r *http.Request) {
		config := WellKnownConfig{
			PolicyDecisionPoint:       baseURL,
			AccessEvaluationEndpoint:  baseURL + "/access/v1/evaluation",
			AccessEvaluationsEndpoint: baseURL + "/access/v1/evaluations",
			SearchSubjectEndpoint:     baseURL + "/access/v1/search/subject",
			SearchResourceEndpoint:    baseURL + "/access/v1/search/resource",
			SearchActionEndpoint:      baseURL + "/access/v1/search/action",
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(config)
	}
}

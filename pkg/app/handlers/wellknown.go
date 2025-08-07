package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/aserto-dev/topaz/pkg/config/v2"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

const AuthZENConfiguration string = `/.well-known/authzen-configuration`

var once sync.Once

type WellKnownConfig struct {
	PolicyDecisionPoint       string `json:"policy_decision_point"`       //nolint: tagliatelle
	AccessEvaluationEndpoint  string `json:"access_evaluation_endpoint"`  //nolint: tagliatelle
	AccessEvaluationsEndpoint string `json:"access_evaluations_endpoint"` //nolint: tagliatelle
	SearchSubjectEndpoint     string `json:"search_subject_endpoint"`     //nolint: tagliatelle
	SearchResourceEndpoint    string `json:"search_resource_endpoint"`    //nolint: tagliatelle
	SearchActionEndpoint      string `json:"search_action_endpoint"`      //nolint: tagliatelle
}

func WellKnownConfigHandler(endpoint *url.URL) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

func SetWellKnownConfigHandler(cfg *config.API, mux *http.ServeMux) error {
	ep, err := endpoint(cfg)
	if err != nil {
		return err
	}

	once.Do(func() {
		mux.HandleFunc(AuthZENConfiguration, WellKnownConfigHandler(ep))
	})

	return nil
}

func endpoint(cfg *config.API) (*url.URL, error) {
	if cfg.Gateway.FQDN != "" {
		return url.Parse(cfg.Gateway.FQDN)
	}

	if cfg.Gateway.ListenAddress != "" {
		u := url.URL{
			Scheme: lo.Ternary(cfg.Gateway.HTTP, "http", "https"),
			Host:   cfg.Gateway.ListenAddress,
		}

		return url.Parse(u.String())
	}

	return nil, errors.Errorf("no fqdn or listen address")
}

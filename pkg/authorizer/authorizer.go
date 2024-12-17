package authorizer

import (
	"encoding/json"
	"time"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/decisionlog"
)

type Config struct {
	RawOPA      map[string]interface{} `json:"opa"`
	OPA         runtime.Config         `json:"-,"`
	DecisionLog decisionlog.Config     `json:"decision_logger"`
	Controller  controller.Config      `json:"controller"`
	JWT         JWTConfig              `json:"jwt"`
}

type JWTConfig struct {
	AcceptableTimeSkew time.Duration `json:"acceptable_time_skew_seconds"`
}

// func (c *Config) OPA() (*runtime.Config, error) {
// 	rCfg := &runtime.Config{}

// 	b, err := json.Marshal(c.RawOPA)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := json.Unmarshal(b, rCfg); err != nil {
// 		return nil, err
// 	}

// 	return rCfg, nil
// }

func (c *Config) UnmarshalJSON(data []byte) error {
	rCfg := runtime.Config{}

	b, err := json.Marshal(c.RawOPA)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &rCfg); err != nil {
		return err
	}

	c.OPA = rCfg

	return nil
}

func (c *Config) MarshalJSON(data []byte) error {
	return nil
}

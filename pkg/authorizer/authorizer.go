package authorizer

import (
	"encoding/json"
	"time"

	"github.com/aserto-dev/aserto-management/controller"
	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/aserto-dev/topaz/pkg/decisionlog"

	"github.com/spf13/viper"
)

type Config struct {
	RawOPA      map[string]interface{} `json:"opa"`
	OPA         runtime.Config         `json:"-,"`
	DecisionLog decisionlog.Config     `json:"decision_logger"`
	Controller  controller.Config      `json:"controller"`
	JWT         JWTConfig              `json:"jwt"`
}

var _ = handler.Config(&Config{})

func (c *Config) SetDefaults(v *viper.Viper, p ...string) {
	c.JWT.SetDefaults(v)
}

func (c *Config) Validate() (bool, error) {
	return true, nil
}

const DefaultAcceptableTimeSkew = time.Second * 5

type JWTConfig struct {
	AcceptableTimeSkew time.Duration `json:"acceptable_time_skew"`
}

func (c *JWTConfig) SetDefaults(v *viper.Viper, p ...string) {
	v.SetDefault("acceptable_time_skew", DefaultAcceptableTimeSkew.String())
}

func (c *JWTConfig) Validate() (bool, error) {
	return true, nil
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

func (c *Config) MarshalJSON() ([]byte, error) {
	return []byte{}, nil
}

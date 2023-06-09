package directory

import (
	grpcc "github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-edge-ds/pkg/directory"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type Config struct {
	Config map[string]interface{} `json:"config"`
}

func (cfg *Config) ToRemoteConfig() (*grpcc.Config, error) {
	grpcCfg := grpcc.Config{}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &grpcCfg,
		TagName: "json",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error decoding file decision logger config")
	}
	err = dec.Decode(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding file decision logger config")
	}

	return &grpcCfg, nil
}

func (cfg *Config) ToEdgeConfig() (*directory.Config, error) {
	edgeCfg := directory.Config{}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &edgeCfg,
		TagName: "json",
	})
	if err != nil {
		return nil, errors.Wrap(err, "error decoding file decision logger config")
	}
	err = dec.Decode(cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding file decision logger config")
	}
	return &edgeCfg, nil
}

package configuration

import (
	"bytes"
	"os"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func LoadConfiguration(fileName string) (*config.Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetConfigFile(fileName)
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	cfg := new(config.Config)

	r := bytes.NewReader(fileContents)
	if err := v.ReadConfig(r); err != nil {
		return nil, err
	}
	err = v.UnmarshalExact(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

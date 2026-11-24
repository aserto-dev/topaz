package topaz

import (
	"sync/atomic"
	"time"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/topaz-opa/internal/config"
	"github.com/aserto-dev/topaz/topaz-opa/internal/plugins/noop"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/util"
)

const defRequestTimeout = 10 * time.Second

type PluginFactory struct{}

var (
	_    plugins.Factory = (*PluginFactory)(nil)
	aCfg atomic.Pointer[Config]
)

func SetConfig(c Config) {
	closeConn()
	aCfg.Store(&c)
}

func GetConfig() Config {
	c := aCfg.Load()
	if c == nil {
		return Config{RequestTimeout: config.Duration(defRequestTimeout)}
	}

	return *c
}

func (p *PluginFactory) New(manager *plugins.Manager, cfg any) plugins.Plugin {
	c, ok := cfg.(Config)
	if !ok || !c.Enabled {
		SetConfig(Config{
			Enabled:        false,
			Connection:     client.Config{},
			RequestTimeout: config.Duration(defRequestTimeout),
		})

		return &noop.Noop{
			Manager: manager,
			Name:    PluginName,
		}
	}

	SetConfig(c)

	return &Plugin{
		config:  &c,
		manager: manager,
	}
}

func (p *PluginFactory) Validate(manager *plugins.Manager, cfg []byte) (any, error) {
	var parsedConfig Config
	if err := util.Unmarshal(cfg, &parsedConfig); err != nil {
		return nil, err
	}

	return parsedConfig, nil
}

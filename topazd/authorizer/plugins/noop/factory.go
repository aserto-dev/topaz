package noop

import (
	"github.com/open-policy-agent/opa/v1/plugins"
)

type PluginFactory struct {
	name string
}

var _ plugins.Factory = (*PluginFactory)(nil)

func NewPluginFactory(name string) PluginFactory {
	return PluginFactory{
		name: name,
	}
}

func (f PluginFactory) New(m *plugins.Manager, config any) plugins.Plugin {
	return &Noop{
		Manager: m,
		Name:    f.name,
	}
}

func (PluginFactory) Validate(m *plugins.Manager, config []byte) (any, error) {
	return map[string]any{}, nil
}

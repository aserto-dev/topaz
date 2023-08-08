package noop

import (
	"github.com/open-policy-agent/opa/plugins"
)

type PluginFactory struct {
	name string
}

func NewPluginFactory(name string) PluginFactory {
	return PluginFactory{
		name: name,
	}
}

func (p PluginFactory) New(m *plugins.Manager, config interface{}) plugins.Plugin {
	return &Noop{
		Manager: m,
		Name:    p.name,
	}
}

func (PluginFactory) Validate(m *plugins.Manager, config []byte) (interface{}, error) {
	return map[string]interface{}{}, nil
}

package handler

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
)

var ErrInvalidConfig = errors.New("invalid plugin configuration")

type Plugin interface {
	IsPlugin()
}

type PluginConfig struct {
	Plugin string `json:"plugin"`
}

func (PluginConfig) IsPlugin() {}

func PluginDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		decodePluginSettings,
	)
}

func decodePluginSettings(from, to reflect.Type, input any) (any, error) {
	if !to.Implements(reflect.TypeOf((*Plugin)(nil)).Elem()) {
		return input, nil
	}

	// This is a plugin. Rename the "settings" field to the value of "plugin".
	v, ok := input.(map[string]any)
	if !ok {
		return input, errors.Wrap(ErrInvalidConfig, "not a yaml mapping")
	}

	plugin, ok := v["plugin"]
	if !ok {
		return input, errors.Wrap(ErrInvalidConfig, "'plugin' field is required")
	}

	pluginName, ok := plugin.(string)
	if !ok {
		return input, errors.Wrap(ErrInvalidConfig, "'plugin' field must be a string")
	}

	settings := v["settings"]

	delete(v, "settings")

	v[pluginName] = settings

	return v, nil
}

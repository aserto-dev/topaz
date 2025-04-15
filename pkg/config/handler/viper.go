package handler

import (
	"maps"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
		// viper.WithDecodeHook(pluginDecodeHook()),
	)

	v.SetConfigType("yaml")

	return v
}

func pluginDecodeHook() mapstructure.DecodeHookFunc {
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

	pluginName, ok := v["plugin"].(string)
	if !ok {
		return input, errors.Wrap(ErrInvalidConfig, "'plugin' field must be a string")
	}

	pluginConfig, ok := v[pluginName].(map[string]any)
	if !ok {
		return input, errors.Wrapf(ErrInvalidConfig, "internal error. plugin '%s' has invalid defaults", pluginName)
	}

	var settings map[string]any

	if s, ok := v["settings"].(map[string]any); ok {
		settings = s
	}

	maps.Copy(pluginConfig, settings)
	delete(v, "settings")

	return v, nil
}

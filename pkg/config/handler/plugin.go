package handler

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

var ErrInvalidConfig = errors.New("invalid plugin configuration")

type Plugin interface {
	IsPlugin()
}

type PluginConfig struct {
	Plugin string `json:"plugin"`
}

func (PluginConfig) IsPlugin() {}

func Indent(s string, n int) string {
	return strings.Join(
		lo.Map(
			strings.Split(strings.TrimSpace(s), "\n"),
			func(line string, _ int) string { return strings.Repeat(" ", n) + line }),
		"\n",
	) + "\n"
}

func PrefixKeys(prefix string, m map[string]any) map[string]any {
	return lo.MapKeys(m, func(_ any, k string) string {
		return prefix + "." + k
	})
}

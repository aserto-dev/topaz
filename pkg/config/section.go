package config

import (
	"io"
	"reflect"
	"strings"

	"github.com/samber/lo"
)

type Section interface {
	Defaults() map[string]any
	Validate() (bool, error)
	Generate(w io.Writer) error
}

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

func WriteIfNotEmpty[T any, P conf[T]](w io.Writer, t *T) error {
	if nilOrEmpty(t) {
		return nil
	}

	return P(t).Generate(w)
}

type conf[T any] interface {
	Section
	*T
}

func nilOrEmpty[T any](t *T) bool {
	if t == nil {
		return true
	}

	var zero T

	return reflect.DeepEqual(zero, *t)
}

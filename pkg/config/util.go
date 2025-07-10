package config

import (
	"bytes"
	"html/template"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/registry/encoding"
	spstr "github.com/go-sprout/sprout/registry/strings"
	"github.com/pborman/indent"
	"github.com/samber/lo"
)

// IndentWriter returns a writer that indents all lines by n spaces.
func IndentWriter(w io.Writer, n int) io.Writer {
	return indent.New(w, strings.Repeat(" ", n))
}

// Indent pads all lines in s by n spaces.
func Indent(s string, n int) string {
	var buf bytes.Buffer

	w := IndentWriter(&buf, n)

	_, _ = w.Write([]byte(s))

	return buf.String()
}

// PrefixKeys adds the given prefix to all keys in the map m.
// A dot is inserted between the prefix and the keys.
// For example:
//
//	PrefixKeys("a.b", map[string]any{"first": 1, "second": 2"})
//
// returns:
//
//	map[string]any{"a.b.first": 1, "a.b.second": 2}
func PrefixKeys(prefix string, m map[string]any) map[string]any {
	return lo.MapKeys(m, func(_ any, k string) string {
		return prefix + "." + k
	})
}

// TemplateFuncs returns a set of commonly used template pipeline functions from
// the github.com/go-sprout/sprout pacakage.
var TemplateFuncs = sync.OnceValue(func() template.FuncMap {
	return sprout.New(sprout.WithRegistries(
		encoding.NewRegistry(),
		spstr.NewRegistry(),
	)).Build()
})

// TrimN removes a leading newline from the given string.
func TrimN(s string) string {
	return strings.TrimPrefix(s, "\n")
}

// WriteNonEmpty serializes t to w if it isn't nil or empty (equal to its default value).
func WriteNonEmpty[T any, P serializer[T]](w io.Writer, t *T) error {
	if nilOrEmpty(t) {
		return nil
	}

	return P(t).Serialize(w)
}

type serializer[T any] interface {
	Serializer
	*T
}

func nilOrEmpty[T any](t *T) bool {
	if t == nil {
		return true
	}

	var zero T

	return reflect.DeepEqual(zero, *t)
}

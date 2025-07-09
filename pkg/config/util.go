package config

import (
	"bytes"
	"html/template"
	"io"
	"reflect"
	"strings"

	"github.com/go-sprout/sprout"
	"github.com/go-sprout/sprout/registry/encoding"
	spstr "github.com/go-sprout/sprout/registry/strings"
	"github.com/pborman/indent"
	"github.com/samber/lo"
)

const yamlIndent = 2

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

var funcs = sprout.New(
	sprout.WithRegistries(encoding.NewRegistry(), spstr.NewRegistry()),
).Build()

// TemplateFuncs returns a set of commonly used template pipeline
// functions.
func TemplateFuncs() template.FuncMap {
	return funcs

	// template.FuncMap{
	// 	"toYAML": func(value any) (string, error) {
	// 		var buf bytes.Buffer
	//
	// 		enc := yaml.NewEncoder(&buf)
	// 		enc.SetIndent(yamlIndent)
	//
	// 		if err := enc.Encode(&value); err != nil {
	// 			return "", errors.Wrap(err, "yaml encoding error")
	// 		}
	//
	// 		return strings.TrimSuffix(buf.String(), "\n"), nil
	// 	},
	// 	"toMap": func(value any) (map[string]any, error) {
	// 		jBytes, err := json.Marshal(value)
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, "json encoding error")
	// 		}
	//
	// 		var m map[string]any
	// 		if err := json.Unmarshal(jBytes, &m); err != nil {
	// 			return nil, errors.Wrap(err, "json encoding error")
	// 		}
	//
	// 		return m, nil
	// 	},
	// )
}

// Trimn removes a leading newline from the given string.
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

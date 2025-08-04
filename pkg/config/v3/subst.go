package config

import (
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	envRegex     = regexp.MustCompile(`(?U:\${(\S+?)})`)
	minVarExpLen = len("${ }")
)

func withEnvSubst(r io.Reader) (io.Reader, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	cfgText := substitueEnvVars(string(body))

	return strings.NewReader(cfgText), nil
}

func substitueEnvVars(s string) string {
	return envRegex.ReplaceAllStringFunc(strings.ReplaceAll(s, `"`, `'`), func(s string) string {
		// Trim off the '${' and '}'
		if len(s) < minVarExpLen {
			// This should never happen..
			return ""
		}

		varName := s[2 : len(s)-1]

		// Lookup the variable in the environment. We play by
		// bash rules.. if its undefined we'll treat it as an
		// empty string instead of raising an error.
		return os.Getenv(varName)
	})
}

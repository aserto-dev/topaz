package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aserto-dev/go-aserto"
	"github.com/pkg/errors"
)

const DefaultRequestTimeout = 5 * time.Second

type Config struct {
	Enabled                 bool          `json:"enabled"`
	Connection              aserto.Config `json:"connection"`
	RequestTimeout          Duration      `json:"request_timeout"`
	EnableDirectoryBuiltIns bool          `json:"enable_directory_builtins"`
	EnableAccessBuiltIns    bool          `json:"enable_access_builtins"`
}

type Duration time.Duration

var ErrInvalidDuration = errors.New("invalid duration format")

func (d *Duration) UnmarshalJSON(b []byte) error {
	// attempt the string form
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		parsed, err := time.ParseDuration(s)
		if err != nil {
			return err
		}

		*d = Duration(parsed)

		return nil
	}

	// fallback to numeric (nanoseconds)
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		*d = Duration(time.Duration(n))
		return nil
	}

	return errors.Wrap(ErrInvalidDuration, string(b))
}

func (d Duration) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%q", time.Duration(d).String())
	return []byte(s), nil
}

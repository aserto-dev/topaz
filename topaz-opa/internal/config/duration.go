package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aserto-dev/topaz/topaz-opa/internal/errs"
)

type Duration time.Duration

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

	return fmt.Errorf("%s %w", string(b), errs.ErrInvalidDuration)
}

func (d Duration) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%q", time.Duration(d).String())
	return []byte(s), nil
}

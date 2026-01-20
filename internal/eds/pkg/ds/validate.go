package ds

import (
	"strings"
)

func IsSet(s string) bool {
	return strings.TrimSpace(s) != ""
}

func IsNotSet(s string) bool {
	return strings.TrimSpace(s) == ""
}

package v3

import (
	"errors"
	"regexp"
)

// Identifier is the string representation of an object, relation and permission type name.
//
// Identifiers are bounded by the underlying defined regex definition (reIdentifier).
//
// An identifier MUST be:
// - have a minimum length of 3 characters
// - have a maximum length of 64 characters
// - start with a character (a-zA-Z)
// - end with a character of a digit (a-zA-Z0-9)
// - can contain dots, underscores and dashes, between the first and last position.

var ErrInvalidIdentifier = errors.New("invalid identifier (" + msgInvalidIdentifier + ")")

//nolint:lll
var (
	reIdentifier         = regexp.MustCompile(`(?m)^[a-zA-Z][a-zA-Z0-9._-]{1,62}[a-zA-Z0-9]$`)
	msgInvalidIdentifier = "must start with a letter, can contain mixed case letters, digits, dots, underscores, and dashes, and must end with a letter or digit"
)

func ValidIdentifier[T interface{ ~string }](in T) bool {
	return reIdentifier.MatchString(string(in))
}

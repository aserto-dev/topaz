package validator

import (
	"regexp"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
)

// TypeIdentifier:
// min length: 3
// max length: 64
// format: ^[a-zA-Z][a-zA-Z0-9._-]{1,62}[a-zA-Z0-9]$
// desc: must start with a letter, can contain mixed case letters, digits, dots, underscores, and dashes, and must end with a letter or digit.
const maxTypeIdentifierLength int = 64

var typeIdentifierMatch = regexp.MustCompile(`(?m)^[a-zA-Z][a-zA-Z0-9._-]{1,62}[a-zA-Z0-9]$`)

func TypeIdentifier(fld Field, val string) error {
	if val == "" {
		return nil
	}

	if len(val) > maxTypeIdentifierLength {
		return derr.ErrTypeIdentifierLength.Msgf(
			"%q: max length of 64 characters exceeded", fld,
		)
	}

	if !typeIdentifierMatch.MatchString(val) {
		return derr.ErrTypeIdentifierFormat.Msgf(
			"%q: must start with a letter, can contain mixed case letters, digits, dots, underscores, and dashes, and must end with a letter or digit", fld,
		)
	}

	return nil
}

// InstanceIdentifier:
// min length: 1
// max length: 256
// format: ^[S]{1,256}$
// desc: cannot contain any spaces or other whitespace characters.
const maxInstanceIdentifierLength int = 256

var instanceIdentifierMatch = regexp.MustCompile(`(?m)^\S{1,256}$`)

func InstanceIdentifier(fld Field, val string) error {
	if val == "" {
		return nil
	}

	if len(val) > maxInstanceIdentifierLength {
		return derr.ErrInstanceIdentifierLength.Msgf(
			"%q: max length of 256 characters exceeded", fld,
		)
	}

	if !instanceIdentifierMatch.MatchString(val) {
		return derr.ErrInstanceIdentifierFormat.Msgf(
			"%q: cannot contain any spaces or other whitespace characters", fld,
		)
	}

	return nil
}

// DisplayName:
// default: ""
// min length: 0
// max length: 100
// format: printable characters.
// desc: must not contain angled brackets, ampersands, or double quotes
const maxDisplayNameLength int = 100

var displayNameMatch = regexp.MustCompile(`(?m)^[[:print:]]{0,100}$`)

func DisplayName(fld Field, val string) error {
	if val == "" {
		return nil
	}

	if len(val) > maxDisplayNameLength {
		return derr.ErrDisplayNameLength.Msgf(
			"%q: max length of 100 characters exceeded", fld,
		)
	}

	if !displayNameMatch.MatchString(val) {
		return derr.ErrDisplayNameFormat.Msgf(
			"%q: can only contain printable characters", fld,
		)
	}

	return nil
}

// Etag:
// default: "0"
// min length: 1
// max length: 20
// format: digits only.
const maxEtagLength int = 20

var etagMatch = regexp.MustCompile(`(?m)^\d{1,20}$`)

func Etag(fld Field, val string) error {
	if val == "" {
		return nil
	}

	if len(val) > maxEtagLength {
		return derr.ErrETagLength.Msgf(
			"%q: max length of 20 characters exceeded", fld,
		)
	}

	if !etagMatch.MatchString(val) {
		return derr.ErrETagFormat.Msgf(
			"%q: can only contain digits", fld,
		)
	}

	return nil
}

// IdentifierTypePresence
// desc: Identifiers always require a type to be specified.
func IdentifierTypePresence(idFld, typeFld Field, idVal, typeVal string) error {
	if idVal != "" && typeVal == "" {
		return derr.ErrMissingTypeIdentifier.Msgf(
			"%q: no type specified for identifier %q, %q must be set", typeFld, idFld, typeFld,
		)
	}

	return nil
}

// Missing checks
// required: true
// ignore_empty: true

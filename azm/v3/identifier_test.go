package v3_test

import (
	"testing"

	v3 "github.com/aserto-dev/topaz/azm/v3"
)

var isValidIdentifierTests = []struct {
	Name  string
	Input string
	Valid bool
}{
	{
		Name:  "invalid empty string",
		Input: "",
		Valid: false,
	},
	{
		Name:  "invalid single character identifier",
		Input: "a",
		Valid: false,
	},
	{
		Name:  "invalid double character identifier",
		Input: "ab",
		Valid: false,
	},
	{
		Name:  "valid triple character identifier abc",
		Input: "abc",
		Valid: true,
	},
	{
		Name:  "valid triple character identifier ab0",
		Input: "ab0",
		Valid: true,
	},
	{
		Name:  "valid triple character identifier a-0",
		Input: "a-0",
		Valid: true,
	},
	{
		Name:  "valid 64 character string",
		Input: "abcd012345678901234567890123456789012345678901234567890123456789",
		Valid: true,
	},
	{
		Name:  "invalid 65 character string",
		Input: "abcde012345678901234567890123456789012345678901234567890123456789",
		Valid: false,
	},
	{
		Name:  "invalid spaces",
		Input: "hello world",
		Valid: false,
	},
	{
		Name:  "invalid uppercase spaces",
		Input: "Hello World 99",
		Valid: false,
	},
	{
		Name:  "invalid start with decimal",
		Input: "10identifier",
		Valid: false,
	},
	{
		Name:  "invalid start with dash",
		Input: "-identifier",
		Valid: false,
	},
	{
		Name:  "invalid start with underscore",
		Input: "_identifier",
		Valid: false,
	},
	{
		Name:  "invalid start with dot",
		Input: ".identifier",
		Valid: false,
	},
	{
		Name:  "valid end with decimal",
		Input: "identifier10",
		Valid: true,
	},
	{
		Name:  "valid embedded underscore end with decimal",
		Input: "identifier_10",
		Valid: true,
	},
	{
		Name:  "valid embedded dash end with decimal",
		Input: "identifier-10",
		Valid: true,
	},
	{
		Name:  "valid embedded dot end with decimal",
		Input: "identifier.10",
		Valid: true,
	},
	{
		Name:  "invalid embedded hash",
		Input: "identifier#10",
		Valid: false,
	},
	{
		Name:  "invalid embedded @ sign",
		Input: "identifier@10",
		Valid: false,
	},
	{
		Name:  "invalid embedded | pipe sign",
		Input: "identifier|10",
		Valid: false,
	},
	{
		Name:  "invalid embedded ? question mark",
		Input: "identifier?10",
		Valid: false,
	},
}

func TestIsValidateIdentifier(t *testing.T) {
	for _, tc := range isValidIdentifierTests {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Valid != v3.ValidIdentifier(tc.Input) {
				t.Fail()
			}
		})
	}
}

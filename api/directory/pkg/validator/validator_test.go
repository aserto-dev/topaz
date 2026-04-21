package validator_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aserto-dev/topaz/api/directory/pkg/derr"
	"github.com/aserto-dev/topaz/api/directory/pkg/validator"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tc struct {
	field    validator.Field
	value    string
	expected error
}

func typeIdentifierTests() []tc {
	return []tc{
		{
			field:    "ident_1",
			value:    "", // no type specified
			expected: nil,
		},
		{
			field:    "ident_2",
			value:    "aaa", // min length
			expected: nil,
		},
		{
			field:    "ident_3",
			value:    "aaa4567890123456678901234566789012345667890123456678901234567890", // max length
			expected: nil,
		},
	}
}

func TestTypeIdentifier(t *testing.T) {
	for i, tc := range typeIdentifierTests() {
		err := validator.TypeIdentifier(tc.field, tc.value)
		if tc.expected == nil {
			assert.NoError(t, err, "test %d", i)
		} else {
			assert.Equal(t, codes.InvalidArgument, status.Convert(err).Code())
		}
	}
}

func instanceIdentifierTests() []tc {
	return []tc{
		{
			field:    "inst_1",
			value:    "aaa", // min length
			expected: nil,
		},
		{
			field:    "inst_2",
			value:    fmt.Sprintf("a23456%s", strings.Repeat("1234567890", 25)), //nolint: perfsprint // max length (256)
			expected: nil,
		},
		{
			field:    "inst_3",
			value:    fmt.Sprintf("a234567%s", strings.Repeat("1234567890", 25)), //nolint: perfsprint // max length exceeded (256+)
			expected: derr.ErrInstanceIdentifierLength,
		},
	}
}

func TestInstanceIdentifier(t *testing.T) {
	for _, tc := range instanceIdentifierTests() {
		err := validator.InstanceIdentifier(tc.field, tc.value)
		if tc.expected == nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, codes.InvalidArgument, status.Convert(err).Code())
		}
	}
}

func displayNameTests() []tc {
	return []tc{
		{
			field:    "dn_1",
			value:    "aaa", // min length
			expected: nil,
		},
		{
			field:    "dn_2",
			value:    strings.Repeat("1234567890", 10), // max length (100)
			expected: nil,
		},
		{
			field:    "dn_3",
			value:    "a" + strings.Repeat("1234567890", 10), // exceeds max length (100+)
			expected: derr.ErrDisplayNameLength,
		},
		{
			field:    "dn_4",
			value:    fmt.Sprintf("%s \t \n ", "Hello World"), //nolint: perfsprint
			expected: nil,
		},
	}
}

func TestDisplayName(t *testing.T) {
	for _, tc := range displayNameTests() {
		err := validator.DisplayName(tc.field, tc.value)
		if tc.expected == nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, codes.InvalidArgument, status.Convert(err).Code())
		}
	}
}

func etagTests() []tc {
	return []tc{
		{
			field:    "etag_1",
			value:    "0", // min length & default value
			expected: nil,
		},
		{
			field:    "etag_2",
			value:    strings.Repeat("1234567890", 2), // max length (20)
			expected: nil,
		},
		{
			field:    "etag_3",
			value:    strings.Repeat("1234567890", 2) + "1", // larger then max length (20+)
			expected: derr.ErrETagLength,
		},
		{
			field:    "etag_4",
			value:    "abc", // no digits
			expected: derr.ErrETagFormat,
		},
		{
			field:    "etag_5",
			value:    "123 456 678", // digits with spaces
			expected: derr.ErrETagFormat,
		},
	}
}

func TestEtag(t *testing.T) {
	for _, tc := range etagTests() {
		err := validator.Etag(tc.field, tc.value)
		if tc.expected == nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, codes.InvalidArgument, status.Convert(err).Code())
		}
	}
}

func TestTypeIdentifierPresence(t *testing.T) {
	assert.NoError(t, validator.IdentifierTypePresence("object_id", "object_type", "", ""))
	assert.NoError(t, validator.IdentifierTypePresence("object_id", "object_type", "", "user"))
	assert.NoError(t, validator.IdentifierTypePresence("object_id", "object_type", "123", "user"))

	assert.ErrorAs(t,
		validator.IdentifierTypePresence("object_id", "object_type", "123", ""),
		&derr.ErrMissingTypeIdentifier,
	)
}

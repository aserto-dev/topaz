package ds_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aserto-dev/topaz/api/directory/v3/reader"
	"github.com/aserto-dev/topaz/topaz-opa/internal/builtins/ds"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errConnect = errors.New("connect failed")

// A client factory failing with an error other than "plugin disabled"
// must be returned as an error, not cause a nil client dereference.
func TestCheckReturnsClientError(t *testing.T) {
	_, fn := ds.RegisterCheck("ds.check", func() (reader.ReaderClient, error) {
		return nil, errConnect
	})

	args := ast.MustParseTerm(`{"object_type": "doc", "object_id": "1", "relation": "viewer", "subject_type": "user", "subject_id": "u1"}`)

	var (
		term *ast.Term
		err  error
	)

	require.NotPanics(t, func() {
		term, err = fn(rego.BuiltinContext{Context: context.Background()}, args)
	})
	assert.Nil(t, term)
	assert.ErrorIs(t, err, errConnect)
}

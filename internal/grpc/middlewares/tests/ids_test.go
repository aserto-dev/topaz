package tests_test

import (
	"context"
	"testing"
	"uuid"

	"github.com/aserto-dev/topaz/internal/headers"
)

func RequestIDContext(t *testing.T) context.Context {
	id := uuid.New()

	return context.WithValue(context.Background(), headers.HeaderAsertoRequestID, id)
}

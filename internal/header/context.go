package header

import (
	"context"

	"github.com/google/uuid"
)

type CtxKey string

var knownValueNames = []CtxKey{
	HeaderAsertoRequestID,
}

func IsValidUUID(uid string) bool {
	_, err := uuid.Parse(uid)
	return err == nil
}

// KnownContextValueStrings is the same as KnownContextValues, but uses string keys (useful for logging).
func KnownContextValueStrings(ctx context.Context) map[string]any {
	result := map[string]any{}

	for _, k := range knownValueNames {
		v := extract(ctx, k)
		if v != "" {
			result[string(k)] = v
		}
	}

	return result
}

func extract(ctx context.Context, key CtxKey) string {
	id, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}

	return id
}

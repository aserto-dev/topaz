package header

import (
	"context"
	"net/textproto"
	"strings"
)

var (
	HeaderAsertoRequestID          = CtxKey(textproto.CanonicalMIMEHeaderKey("Aserto-Request-Id"))
	HeaderAsertoRequestIDLowercase = CtxKey(strings.ToLower(string(HeaderAsertoRequestID)))
)

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, HeaderAsertoRequestID, requestID)
}

func ExtractRequestID(ctx context.Context) string {
	id, ok := ctx.Value(HeaderAsertoRequestID).(string)
	if !ok {
		return ""
	}

	return id
}

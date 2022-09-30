package instance

import (
	"context"

	"github.com/rs/zerolog"
)

func GetInstanceLogger(ctx context.Context, log *zerolog.Logger) *zerolog.Logger {
	instanceLogger := log.With().Fields(ExtractID(ctx)).Logger()
	return &instanceLogger
}

func ExtractID(ctx context.Context) string {
	id, ok := ctx.Value(InstanceIDHeader).(string)
	if !ok {
		return ""
	}

	return id
}

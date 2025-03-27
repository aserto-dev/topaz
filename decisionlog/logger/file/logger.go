package file

import (
	"context"
	"encoding/json"

	api "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/decisionlog"
	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type fileLogger zerolog.Logger

var _ decisionlog.DecisionLogger = (*fileLogger)(nil)

func New(ctx context.Context, cfg *Config, logger *zerolog.Logger) (*fileLogger, error) {
	cfg.SetDefaults()

	ljLogger := &lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    cfg.MaxFileSizeMB,
		MaxBackups: cfg.MaxFileCount,
	}

	decisionLogger := zerolog.New(ljLogger)

	return (*fileLogger)(&decisionLogger), nil
}

func (l *fileLogger) Log(d *api.Decision) error {
	bytes, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "error marshaling decision")
	}

	(*zerolog.Logger)(l).Log().Msg(string(bytes))

	return nil
}

func (l *fileLogger) Shutdown() {
}

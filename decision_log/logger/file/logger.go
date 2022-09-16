package file

import (
	"context"
	"encoding/json"

	api "github.com/aserto-dev/go-authorizer/aserto/api/v2"
	decisionlog "github.com/aserto-dev/topaz/decision_log"
	"github.com/pkg/errors"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type fileLogger zerolog.Logger

func New(ctx context.Context, cfg *Config, logger *zerolog.Logger) (decisionlog.DecisionLogger, error) {
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

package logger

import (
	"maps"
	"sync"

	"github.com/open-policy-agent/opa/v1/logging"
	"github.com/rs/zerolog"
)

type OpaLogger struct {
	logger    *zerolog.Logger
	fields    map[string]any
	levelLock sync.Mutex
}

var _ logging.Logger = &OpaLogger{}

func NewOpaLogger(logger *zerolog.Logger) *OpaLogger {
	return &OpaLogger{logger: logger}
}

func (l *OpaLogger) Debug(fmt string, a ...any) {
	l.logger.Debug().Msgf(fmt, a...)
}

func (l *OpaLogger) Info(fmt string, a ...any) {
	l.logger.Info().Msgf(fmt, a...)
}

func (l *OpaLogger) Error(fmt string, a ...any) {
	l.logger.Error().Msgf(fmt, a...)
}

func (l *OpaLogger) Warn(fmt string, a ...any) {
	l.logger.Warn().Msgf(fmt, a...)
}

func (l *OpaLogger) WithFields(fields map[string]any) logging.Logger {
	newLogger := l.logger.With().Fields(fields).Logger()
	logger := NewOpaLogger(&newLogger)
	logger.fields = make(map[string]any)

	maps.Copy(logger.fields, l.fields)

	maps.Copy(logger.fields, fields)

	return logger
}

func (l *OpaLogger) GetFields() map[string]any {
	return l.fields
}

func (l *OpaLogger) GetLevel() logging.Level {
	switch l.logger.GetLevel() { //nolint:exhaustive
	case zerolog.DebugLevel:
		return logging.Debug
	case zerolog.InfoLevel:
		return logging.Info
	case zerolog.WarnLevel:
		return logging.Warn
	case zerolog.ErrorLevel:
		return logging.Error
	default:
		return logging.Error
	}
}

func (l *OpaLogger) SetLevel(level logging.Level) {
	zLevel := zerolog.DebugLevel

	switch level {
	case logging.Debug:
		zLevel = zerolog.DebugLevel
	case logging.Info:
		zLevel = zerolog.InfoLevel
	case logging.Warn:
		zLevel = zerolog.WarnLevel
	case logging.Error:
		zLevel = zerolog.ErrorLevel
	}

	l.levelLock.Lock()
	defer l.levelLock.Unlock()

	newLogger := l.logger.Level(zLevel)
	l.logger = &newLogger
}

package logger

import (
	"log"
	"sync"

	"github.com/rs/zerolog"
)

// ZerologWriter implements io.Writer for a zerolog logger.
type ZerologWriter struct {
	logger *zerolog.Logger
	level  zerolog.Level
}

var stdLocker = &sync.Mutex{}

// NewZerologWriter creates a new ZerologWriter.
func NewZerologWriter(logger *zerolog.Logger) *ZerologWriter {
	return NewZerologWriterWithLevel(logger, zerolog.InfoLevel)
}

func NewZerologWriterWithLevel(logger *zerolog.Logger, level zerolog.Level) *ZerologWriter {
	return &ZerologWriter{logger: logger, level: level}
}

func (z *ZerologWriter) Write(p []byte) (int, error) {
	msg := string(p)

	stdLocker.Lock()
	defer stdLocker.Unlock()

	switch z.level { //nolint:exhaustive
	case zerolog.InfoLevel:
		z.logger.Info().Msg(msg)
	case zerolog.DebugLevel:
		z.logger.Debug().Msg(msg)
	case zerolog.ErrorLevel:
		z.logger.Error().Msg(msg)
	case zerolog.WarnLevel:
		z.logger.Warn().Msg(msg)
	case zerolog.TraceLevel:
		z.logger.Trace().Msg(msg)
	default:
		z.logger.Info().Msg(msg)
	}

	return len(p), nil
}

// NewSTDLogger creates a standard logger that writes to
// a zerolog logger.
func NewSTDLogger(logger *zerolog.Logger) *log.Logger {
	return log.New(NewZerologWriter(logger), "", 0)
}

package logger

import (
	"fmt"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/grpclog"
)

// GRPCZeroLogger is a GRPC logger that uses zerolog.
type GRPCZeroLogger struct {
	log   *zerolog.Logger
	level zerolog.Level
}

var _ grpclog.LoggerV2 = GRPCZeroLogger{}

func SetGRPCLogger(log *zerolog.Logger, level zerolog.Level) {
	grpclog.SetLoggerV2(NewGRPCZeroLogger(log, level))
}

// NewGRPCZeroLogger creates a GRPCZeroLogger.
func NewGRPCZeroLogger(log *zerolog.Logger, level zerolog.Level) GRPCZeroLogger {
	log.Debug().Msgf("GRPC log level is: %s", level.String())
	grpcLogger := log.Level(level).With().Str("log-source", "grpc").Logger()

	return GRPCZeroLogger{log: &grpcLogger, level: level}
}

// Fatal logs a fatal message.
func (l GRPCZeroLogger) Fatal(args ...any) {
	l.log.Fatal().Msg(fmt.Sprint(args...))
}

// Fatalf formats and logs a fatal message.
func (l GRPCZeroLogger) Fatalf(format string, args ...any) {
	l.log.Fatal().Msg(fmt.Sprintf(format, args...))
}

// Fatalln logs a fatal message and a newline.
func (l GRPCZeroLogger) Fatalln(args ...any) {
	l.Fatal(args...)
}

// Error logs an error message.
func (l GRPCZeroLogger) Error(args ...any) {
	l.log.Error().Msg(fmt.Sprint(args...))
}

// Errorf formats and logs an error message.
func (l GRPCZeroLogger) Errorf(format string, args ...any) {
	l.log.Error().Msg(fmt.Sprintf(format, args...))
}

// Errorln logs an error message and a newline.
func (l GRPCZeroLogger) Errorln(args ...any) {
	l.Error(args...)
}

// Info logs an info message.
func (l GRPCZeroLogger) Info(args ...any) {
	l.log.Info().Msg(fmt.Sprint(args...))
}

// Infof formats and logs an info message.
func (l GRPCZeroLogger) Infof(format string, args ...any) {
	l.log.Info().Msg(fmt.Sprintf(format, args...))
}

// Infoln formats and logs an info message and a newline.
func (l GRPCZeroLogger) Infoln(args ...any) {
	l.Info(args...)
}

// Warning logs a warning message.
func (l GRPCZeroLogger) Warning(args ...any) {
	l.log.Warn().Msg(fmt.Sprint(args...))
}

// Warningf formats and logs a warning message.
func (l GRPCZeroLogger) Warningf(format string, args ...any) {
	l.log.Warn().Msg(fmt.Sprintf(format, args...))
}

// Warningln formats and logs a warning message and a newline.
func (l GRPCZeroLogger) Warningln(args ...any) {
	l.Warning(args...)
}

// Print prints a message.
func (l GRPCZeroLogger) Print(args ...any) {
	l.log.Print(args...)
}

// Printf formats and prints a message.
func (l GRPCZeroLogger) Printf(format string, args ...any) {
	l.log.Printf(format, args...)
}

// Println prints a message and a newline.
func (l GRPCZeroLogger) Println(args ...any) {
	l.Print(args...)
}

// V always returns true.
func (l GRPCZeroLogger) V(_ int) bool {
	return true
}

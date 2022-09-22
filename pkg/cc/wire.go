//go:build wireinject
// +build wireinject

package cc

import (
	"github.com/aserto-dev/aserto-certs/certs"
	logger "github.com/aserto-dev/aserto-logger"
	"github.com/google/wire"

	"github.com/aserto-dev/topaz/pkg/cc/config"
	cc_context "github.com/aserto-dev/topaz/pkg/cc/context"
)

var (
	ccSet = wire.NewSet(
		cc_context.NewContext,
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		certs.NewGenerator,

		wire.Struct(new(CC), "*"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup"),
	)

	ccTestSet = wire.NewSet(
		// Test
		cc_context.NewTestContext,

		// Normal
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		certs.NewGenerator,

		wire.Struct(new(CC), "*"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup"),
	)
)

// buildCC sets up the CC struct that contains all dependencies that
// are cross cutting
func buildCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	wire.Build(ccSet)
	return &CC{}, func() {}, nil
}

func buildTestCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	wire.Build(ccTestSet)
	return &CC{}, func() {}, nil
}

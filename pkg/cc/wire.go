//go:build wireinject
// +build wireinject

package cc

import (
	"github.com/aserto-dev/certs"
	"github.com/aserto-dev/logger"
	"github.com/google/wire"

	runtimeLogger "github.com/aserto-dev/runtime/logger"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	cc_context "github.com/aserto-dev/topaz/pkg/cc/context"
)

var (
	commonSet = wire.NewSet(
		config.NewConfig,
		config.NewLoggerConfig,
		runtimeLogger.NewLogger,
		certs.NewGenerator,

		wire.Struct(new(CC), "*"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup"),
	)

	ccSet = wire.NewSet(
		commonSet,
		cc_context.NewContext,
	)

	ccTestSet = wire.NewSet(
		commonSet,
		cc_context.NewTestContext,
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

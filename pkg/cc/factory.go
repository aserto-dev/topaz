package cc

import (
	"github.com/aserto-dev/certs"
	logger "github.com/aserto-dev/logger"
	runtime_logger "github.com/aserto-dev/runtime/logger"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/aserto-dev/topaz/pkg/cc/context"
)

// buildCC sets up the CC struct that contains all dependencies that are cross cutting.
func buildCC(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (
	*CC,
	func(),
	error,
) {
	errGroupAndContext := context.NewContext()
	contextContext := errGroupAndContext.Ctx

	loggerConfig, err := config.NewLoggerConfig(configPath, overrides)
	if err != nil {
		return nil, nil, err
	}

	zerologLogger, err := runtime_logger.NewLogger(logOutput, errOutput, loggerConfig)
	if err != nil {
		return nil, nil, err
	}

	generator := certs.NewGenerator(zerologLogger)

	configConfig, err := config.NewConfig(configPath, zerologLogger, overrides, generator)
	if err != nil {
		return nil, nil, err
	}

	group := errGroupAndContext.ErrGroup
	ccCC := &CC{
		Context:  contextContext,
		Config:   configConfig,
		Log:      zerologLogger,
		ErrGroup: group,
	}

	return ccCC, func() {
	}, nil
}

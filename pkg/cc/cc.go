package cc

import (
	"context"
	"sync"

	"github.com/aserto-dev/logger"
	"github.com/aserto-dev/topaz/pkg/cc/config"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// CC contains dependencies that are cross cutting and are needed in most
// of the providers that make up this application.
type CC struct {
	Context  context.Context
	Config   *config.Config
	Log      *zerolog.Logger
	ErrGroup *errgroup.Group
}

var (
	once         sync.Once
	cc           *CC
	cleanup      func()
	errSingleton error
)

// NewCC creates a singleton CC.
func NewCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, errSingleton = buildCC(logOutput, errOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()
		once = sync.Once{}
	}, errSingleton
}

// NewTestCC creates a singleton CC to be used for testing.
// It uses a fake context (context.Background).
func NewTestCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, errSingleton = buildTestCC(logOutput, errOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()
		once = sync.Once{}
	}, errSingleton
}

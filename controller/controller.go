package controller

import (
	"context"
	"math"
	"time"

	api "github.com/aserto-dev/go-grpc/aserto/api/v2"
	management "github.com/aserto-dev/go-grpc/aserto/management/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type CommandFunc func(context.Context, *api.Command) error

type Controller struct {
	client       management.ControllerClient
	instanceInfo *api.InstanceInfo
	logger       *zerolog.Logger
	handler      CommandFunc
}

type sleepResult bool

const (
	Canceled        sleepResult = true
	DurationReached sleepResult = false
)

const (
	timeout    = 1 * time.Second
	maxBackoff = 600 * time.Second // max waits between retries.
)

func NewController(logger *zerolog.Logger, policyName, host string, cfg *Config, handler CommandFunc) (*Controller, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	newLogger := logger.With().Fields(map[string]interface{}{
		"component":   "controller",
		"tenant-id":   cfg.Server.TenantID,
		"policy-name": policyName,
		"host":        host,
	}).Logger()

	conn, err := cfg.Server.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize new connection")
	}

	return &Controller{
		client: management.NewControllerClient(conn),
		instanceInfo: &api.InstanceInfo{
			PolicyName:  policyName,
			PolicyLabel: policyName,
			RemoteHost:  host,
		},
		logger:  &newLogger,
		handler: handler,
	}, nil
}

// run context dependant controller.
func (c *Controller) Start(ctx context.Context) func() {
	ctx, cancel := context.WithCancel(ctx)
	errGroup := errgroup.Group{}

	errGroup.Go(func() error {
		c.logger.Trace().Msg("command loop running")

		for retry := 0; ; retry++ {
			retry++

			err := c.runCommandLoop(ctx)
			if err == nil { // graceful shutdown on context canceled.
				c.logger.Trace().Msg("command loop ended")
				return nil
			}

			c.logger.Info().Err(err).Msg("command loop exited with error, restarting")

			backoff := timeout * time.Duration(math.Pow(2, float64(retry)))
			if sleepWithContext(ctx, min(backoff, maxBackoff)) == Canceled {
				c.logger.Trace().Msg("command loop canceled")
				break
			}
		}

		return nil
	})

	return func() {
		cancel()

		if err := errGroup.Wait(); err != nil {
			c.logger.Error().Err(err).Msg("cleanup error")
		}
	}
}

func (c *Controller) runCommandLoop(ctx context.Context) error {
	stream, err := c.client.CommandStream(ctx, &management.CommandStreamRequest{
		Info: c.instanceInfo,
	})
	if err != nil {
		return errors.Wrap(err, "failed to establish command stream with control plane")
	}

	errGroup := errgroup.Group{}
	errGroup.Go(func() error {
		for {
			cmd, errRcv := stream.Recv()
			if errRcv != nil {
				return errRcv
			}

			c.logger.Debug().Msgf("processing remote command %v", cmd.Command)

			if err := c.handler(context.Background(), cmd.Command); err != nil {
				c.logger.Error().Err(err).Msg("error processing command")
			}

			c.logger.Trace().Msg("successfully processed remote command")
		}
	})

	return errGroup.Wait()
}

func sleepWithContext(ctx context.Context, duration time.Duration) sleepResult {
	select {
	case <-ctx.Done():
		return Canceled
	case <-time.After(duration):
		return DurationReached
	}
}

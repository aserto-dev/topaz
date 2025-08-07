package controller

import (
	"context"
	"math"
	"time"

	client "github.com/aserto-dev/go-aserto"
	api "github.com/aserto-dev/go-grpc/aserto/api/v2"
	management "github.com/aserto-dev/go-grpc/aserto/management/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type CommandFunc func(context.Context, *api.Command) error

type Controller struct {
	client       management.ControllerClient
	instanceInfo *api.InstanceInfo
	logger       *zerolog.Logger
	handler      CommandFunc
	errGroup     errgroup.Group
	cancel       func()
}

type sleepResult bool

const (
	Canceled        sleepResult = true
	DurationReached sleepResult = false
)

const (
	timeout          = 1 * time.Second
	maxBackoff       = 600 * time.Second // max waits between retries.
	keepaliveTime    = 30 * time.Second  // send pings every 30 seconds if there is no activity
	keepaliveTimeout = 5 * time.Second   // wait 5 seconds for ping ack before considering the connection dead

)

func NewController(logger *zerolog.Logger, policyName, host string, cfg *Config, handler CommandFunc) (*Controller, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	newLogger := logger.With().Fields(map[string]any{
		"component":   "controller",
		"tenant-id":   cfg.Server.TenantID,
		"policy-name": policyName,
		"host":        host,
	}).Logger()

	opts := []grpc.DialOption{grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    keepaliveTime,    // send pings every 30 seconds if there is no activity
		Timeout: keepaliveTimeout, // wait 5 seconds for ping ack before considering the connection dead
	})}

	conn, err := cfg.Server.Connect(client.WithDialOptions(opts...))
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
		cancel:  func() {},
	}, nil
}

const expBackoff = 2

// run context dependant controller.
func (c *Controller) Start(ctx context.Context) {
	if c.client == nil {
		return
	}

	ctx, cancel := context.WithCancel(ctx)

	c.errGroup.Go(func() error {
		c.logger.Trace().Msg("command loop running")

		for retry := 0; ; retry++ {
			retry++

			err := c.runCommandLoop(ctx)
			if err == nil { // graceful shutdown on context canceled.
				c.logger.Trace().Msg("command loop ended")
				return nil
			}

			c.logger.Info().Err(err).Msg("command loop exited with error, restarting")

			backoff := timeout * time.Duration(math.Pow(expBackoff, float64(retry)))
			if sleepWithContext(ctx, min(backoff, maxBackoff)) == Canceled {
				c.logger.Trace().Msg("command loop canceled")
				break
			}
		}

		return nil
	})

	c.cancel = cancel
}

func (c *Controller) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}

	return c.errGroup.Wait()
}

func (c *Controller) runCommandLoop(ctx context.Context) error {
	stream, err := c.client.CommandStream(ctx, &management.CommandStreamRequest{
		Info: c.instanceInfo,
	})
	if err != nil {
		return errors.Wrap(err, "failed to establish command stream with control plane")
	}

	errs := make(chan error, 1)

	go func() {
		for {
			cmd, errRcv := stream.Recv()
			if errRcv != nil {
				errs <- errRcv
				return
			}

			c.logger.Debug().Msgf("processing remote command %v", cmd.GetCommand())

			if err := c.handler(context.Background(), cmd.GetCommand()); err != nil {
				c.logger.Error().Err(err).Msg("error processing command")
			}

			c.logger.Trace().Msg("successfully processed remote command")
		}
	}()

	return <-errs
}

func sleepWithContext(ctx context.Context, duration time.Duration) sleepResult {
	select {
	case <-ctx.Done():
		return Canceled
	case <-time.After(duration):
		return DurationReached
	}
}

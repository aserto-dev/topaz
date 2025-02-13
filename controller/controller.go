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
	ctx        context.Context
	client     management.ControllerClient
	policyName string
	host       string
	cfg        *Config
	logger     *zerolog.Logger
	handle     CommandFunc
}

type sleepResult bool

const (
	timeout                     = 1 * time.Second
	maxBackoff                  = 600 * time.Second // max waits between retries.
	Canceled        sleepResult = true
	DurationReached sleepResult = false
)

func NewController(logger *zerolog.Logger, ctx context.Context, policyName, host string, cfg *Config, handler CommandFunc) (*Controller, error) {
	if cfg.Server == nil {
		return nil, errors.New("no server configuration provided")
	}
	if !cfg.Enabled {
		return nil, nil
	}
	conn, err := cfg.Server.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize new connection")
	}
	remoteCli := management.NewControllerClient(conn)
	newLogger := logger.With().Str("component", "controller").Logger()

	return &Controller{
		ctx:        ctx,
		client:     remoteCli,
		policyName: policyName,
		host:       host,
		cfg:        cfg,
		logger:     &newLogger,
		handle:     handler,
	}, nil
}

// run context dependant controller.
func (c *Controller) Start() func() {
	logger := c.logger.With().Fields(map[string]interface{}{
		"tenant-id":   c.cfg.Server.TenantID,
		"policy-name": c.policyName,
		"host":        c.host,
	}).Logger()

	ctx, cancel := context.WithCancel(c.ctx)
	errGroup := errgroup.Group{}

	errGroup.Go(func() error {
		for retry := 0; ; retry++ {
			retry++
			err := c.runCommandLoop()
			if err == nil { // graceful shutdown on context canceled.
				return nil
			}
			logger.Info().Err(err).Msg("command loop exited with error, restarting")
			backoff := timeout * time.Duration(math.Pow(2, float64(retry)))
			if sleepWithContext(ctx, min(backoff, maxBackoff)) == Canceled {
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

func (c *Controller) runCommandLoop() error {
	stream, err := c.client.CommandStream(c.ctx, &management.CommandStreamRequest{
		Info: &api.InstanceInfo{

			PolicyName:  c.policyName,
			PolicyLabel: c.policyName,
			RemoteHost:  c.host,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to establish command stream with control plane")
	}

	errCh := make(chan error)
	go func() {
		for {
			cmd, errRcv := stream.Recv()
			if errRcv != nil {
				errCh <- errRcv
				return
			}

			c.logger.Trace().Msg("processing remote command")
			err := c.handle(context.Background(), cmd.Command)
			if err != nil {
				c.logger.Error().Err(err).Msg("error processing command")
			}
			c.logger.Trace().Msg("successfully processed remote command")
		}
	}()

	c.logger.Trace().Msg("command loop running")
	defer func() {
		c.logger.Trace().Msg("command loop ended")
	}()

	select {
	case err = <-errCh:
		c.logger.Info().Err(err).Msg("error receiving command")
		return err
	case <-stream.Context().Done():
		c.logger.Trace().Msg("stream context done")
		return stream.Context().Err()
	case <-c.ctx.Done():
		c.logger.Trace().Msg("context done")
		return nil
	}
}

func sleepWithContext(ctx context.Context, duration time.Duration) sleepResult {
	select {
	case <-ctx.Done():
		return Canceled
	case <-time.After(duration):
		return DurationReached
	}
}

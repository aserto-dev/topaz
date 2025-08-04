package controller_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	client "github.com/aserto-dev/go-aserto"
	api "github.com/aserto-dev/go-grpc/aserto/api/v2"
	"github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitializeErrorController(t *testing.T) {
	logger := zerolog.Nop()

	ctrl, err := controller.NewController(
		&logger, "test", "test-host",
		&controller.Config{
			Optional: config.Optional{
				Enabled: true,
			},
		},
		func(ctx context.Context, c *api.Command) error {
			return nil
		},
	)

	require.Error(t, err, "no server configuration provided")
	assert.Nil(t, ctrl)
}

func TestNotEnabledController(t *testing.T) {
	logger := zerolog.Nop()
	ctrl, err := controller.NewController(
		&logger, "test", "test-host",
		&controller.Config{
			Optional: config.Optional{
				Enabled: false,
			},
		},
		func(ctx context.Context, c *api.Command) error {
			return nil
		},
	)

	require.NoError(t, err)
	assert.Nil(t, ctrl)
}

func TestEnabledController(t *testing.T) {
	logger := zerolog.Nop()
	ctrl, err := controller.NewController(
		&logger, "test", "test-host",
		&controller.Config{
			Optional: config.Optional{
				Enabled: true,
			},
			Server: client.Config{
				Address: "localhost:1234",
			},
		},
		func(ctx context.Context, c *api.Command) error {
			return nil
		},
	)

	require.NoError(t, err)
	assert.NotNil(t, ctrl)
}

func TestControllerLogMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: buf, NoColor: true})
	ctrl, err := controller.NewController(
		&logger, "test", "test-host",
		&controller.Config{
			Optional: config.Optional{
				Enabled: true,
			},
			Server: client.Config{
				Address: "localhost:1234",
			},
		},
		func(ctx context.Context, c *api.Command) error {
			return nil
		},
	)

	require.NoError(t, err)
	assert.NotNil(t, ctrl)

	ctrl.Start(t.Context())

	time.Sleep(1 * time.Second)

	defer ctrl.Stop()

	logMessages := buf.String()

	t.Log(logMessages)

	assert.Contains(t, logMessages, "command loop exited with error, restarting")
	assert.Contains(t, logMessages, "component=controller")
	assert.Contains(t, logMessages, "host=test-host")
}

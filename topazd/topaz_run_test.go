package main

import (
	"context"
	"testing"

	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/stretchr/testify/require"
)

func TestResolveRunConfigFileExplicit(t *testing.T) {
	actual, err := resolveRunConfigFile(context.Background(), "/config/topaz.yaml", false)

	require.NoError(t, err)
	require.Equal(t, "/config/topaz.yaml", actual)
}

func TestResolveRunConfigFileRequiresConfigOutsideDebugMode(t *testing.T) {
	_, err := resolveRunConfigFile(context.Background(), "", false)

	require.EqualError(t, err, "config-file is required")
}

func TestResolveRunConfigFileFromCLIConfig(t *testing.T) {
	originalLoadCLIConfig := loadCLIConfig
	t.Cleanup(func() { loadCLIConfig = originalLoadCLIConfig })

	loadCLIConfig = func(context.Context) (*cc.Config, error) {
		return &cc.Config{
			Active:  cc.ActiveConfig{ConfigFile: "/config/active.yaml"},
			Running: cc.RunningConfig{ActiveConfig: cc.ActiveConfig{ConfigFile: "/config/running.yaml"}},
		}, nil
	}

	actual, err := resolveRunConfigFile(context.Background(), "", true)

	require.NoError(t, err)
	require.Equal(t, "/config/running.yaml", actual)
}

func TestResolveRunConfigFileFallsBackToActiveConfig(t *testing.T) {
	originalLoadCLIConfig := loadCLIConfig
	t.Cleanup(func() { loadCLIConfig = originalLoadCLIConfig })

	loadCLIConfig = func(context.Context) (*cc.Config, error) {
		return &cc.Config{Active: cc.ActiveConfig{ConfigFile: "/config/active.yaml"}}, nil
	}

	actual, err := resolveRunConfigFile(context.Background(), "", true)

	require.NoError(t, err)
	require.Equal(t, "/config/active.yaml", actual)
}

package runtime

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/open-policy-agent/opa/v1/metrics"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/open-policy-agent/opa/v1/plugins/discovery"
	"github.com/open-policy-agent/opa/v1/plugins/status"
	"github.com/pkg/errors"
)

const (
	bundleErrorCode     = "bundle_error"
	discoveryPluginName = "discovery"
	bundlePluginName    = "bundle"
	statusPluginName    = status.Name
	pluginWait          = 10 * time.Millisecond
)

type PluginDefinition struct {
	Name    string
	Factory plugins.Factory
}

// WaitForPlugins waits for all plugins to be ready.
func (r *Runtime) WaitForPlugins(timeoutCtx context.Context, maxWaitTime time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(timeoutCtx, maxWaitTime)
	defer cancel()

	for {
		s := r.Status()
		if s.Ready {
			r.Logger.Info().Msg("runtime is ready")
			return nil
		}

		errs := s.Errors

		for i := range s.Bundles {
			errs = append(errs, s.Bundles[i].Errors...)
		}

		if len(errs) > 0 {
			return errors.Wrap(multierror.Append(nil, errs...), "error loading plugins")
		}

		if timeoutCtx.Err() != nil {
			return errors.Wrap(timeoutCtx.Err(), "timeout while waiting for runtime to load")
		}

		time.Sleep(pluginWait)
	}
}

type bundleState struct {
	revision       string
	errors         []error
	message        string
	metrics        metrics.Metrics
	lastActivation time.Time
	lastDownload   time.Time
}

type pluginState struct {
	err    error
	loaded bool
}

// pluginsLoaded returns true if all plugins have been loaded.
func (r *Runtime) pluginsLoaded() bool {
	if r.pluginsManager == nil {
		return false
	}

	pluginStates := r.pluginsManager.PluginStatus()
	for pluginName, status := range pluginStates {
		if status == nil || status.State == plugins.StateOK {
			continue
		}

		if pluginName == discoveryPluginName && r.Config.Config.Discovery == nil {
			continue
		}

		if pluginName == statusPluginName {
			continue
		}

		r.Logger.Trace().Str("state", string(status.State)).Str("plugin-name", pluginName).Msg("plugin not ready")

		return false
	}

	return true
}

func (r *Runtime) bundlesStatusCallback(status bundle.Status) {
	errs := status.Errors
	if status.Code == bundleErrorCode {
		errs = append(errs, errors.Errorf("bundle error: %s", status.Message))
	}

	r.bundleStates.Store(status.Name, &bundleState{
		revision:       status.ActiveRevision,
		errors:         errs,
		message:        status.Message,
		metrics:        status.Metrics,
		lastActivation: status.LastSuccessfulActivation,
		lastDownload:   status.LastSuccessfulDownload,
	})
	r.setLatestStatus(r.status())
}

func (r *Runtime) pluginStatusCallback(statusDetails map[string]*plugins.Status) {
	for n, s := range statusDetails {
		if n == bundlePluginName && !r.bundlesCallbackRegistered.Load() {
			r.registerBundlesCallback()
		}

		if n == discoveryPluginName && !r.discoveryCallbackRegistered.Load() {
			r.registerDiscoveryCallback()
		}

		if s == nil {
			continue
		}

		switch s.State {
		case plugins.StateErr:
			r.Logger.Trace().Str("runtime-id", r.pluginsManager.ID).Str("plugin", n).Msg("plugin in error state")
			r.pluginStates.Store(n, &pluginState{err: errors.New("there was an error loading the plugin"), loaded: false})

		case plugins.StateNotReady:
			r.Logger.Trace().Str("runtime-id", r.pluginsManager.ID).Str("plugin", n).Msg("plugin not ready")
			r.pluginStates.Store(n, &pluginState{loaded: false})

		case plugins.StateWarn:
			r.Logger.Trace().Str("runtime-id", r.pluginsManager.ID).Str("plugin", n).Msg("plugin in warning state")
			r.pluginStates.Store(n, &pluginState{loaded: true})

		case plugins.StateOK:
			r.Logger.Trace().Str("runtime-id", r.pluginsManager.ID).Str("plugin", n).Msg("plugin ready")
			r.pluginStates.Store(n, &pluginState{loaded: true})
		}
	}

	r.setLatestStatus(r.status())
}

func (r *Runtime) registerBundlesCallback() {
	if plugin := r.pluginsManager.Plugin(bundlePluginName); plugin != nil {
		bundlePlugin, ok := plugin.(*bundle.Plugin)
		if !ok {
			r.Logger.Error().Type("plugin", plugin).Msg("unexpected bundle plugin type")
			return
		}

		bundlePlugin.Register("aserto-error-recorder", r.bundlesStatusCallback)
		r.bundlesCallbackRegistered.Store(true)
	}
}

func (r *Runtime) registerDiscoveryCallback() {
	if plugin := r.pluginsManager.Plugin(discoveryPluginName); plugin != nil {
		discoveryPlugin, ok := plugin.(*discovery.Discovery)
		if !ok {
			r.Logger.Error().Type("plugin", plugin).Msg("unexpected discovery plugin type")
			return
		}

		discoveryPlugin.RegisterListener("aserto-error-recorder", r.bundlesStatusCallback)
		r.discoveryCallbackRegistered.Store(true)
	}
}

package authorizer

import (
	"context"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/aserto-dev/go-grpc/aserto/api/v2"
	rt "github.com/aserto-dev/runtime"
	"github.com/open-policy-agent/opa/v1/plugins/discovery"

	ctrl "github.com/aserto-dev/topaz/controller"
	"github.com/aserto-dev/topaz/pkg/cli/x"
	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/plugins/edge"
)

type ControllerConfig ctrl.Config

var _ config.Section = (*ControllerConfig)(nil)

func (c *ControllerConfig) Defaults() map[string]any {
	return map[string]any{}
}

func (c *ControllerConfig) Validate() error {
	return nil
}

func (c *ControllerConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("CONTROLLER").Parse(controllerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const controllerTemplate = `
# control plane configuration
controller:
  enabled: {{ .Enabled }}
  {{- if .Enabled }}
  server:
    address: '{{ .Server.Address }}'
    api_key: '{{ .Server.APIKey }}'
    client_cert_path: '{{ .Server.ClientCertPath }}'
    client_key_path: '{{ .Server.ClientKeyPath }}'
  {{ end }}
`

func newController(cfg *Config, logger *zerolog.Logger, runtime *rt.Runtime) (*ctrl.Controller, error) {
	if cfg.OPA.Config.Discovery == nil {
		return &ctrl.Controller{}, nil
	}

	host, err := hostname()
	if err != nil {
		return nil, err
	}

	if cfg.OPA.Config.Discovery.Resource == nil {
		return nil, aerr.ErrBadRuntime.Msg("discovery resource must be provided")
	}

	details := strings.Split(*cfg.OPA.Config.Discovery.Resource, "/")

	if cfg.Controller.Server.TenantID == "" {
		cfg.Controller.Server.TenantID = cfg.OPA.InstanceID // get the tenant id from the opa instance id config.
	}

	if len(details) < 1 {
		return nil, aerr.ErrBadRuntime.Msg("provided discovery resource not formatted correctly")
	}

	return ctrl.NewController(
		logger,
		details[0],
		host,
		(*ctrl.Config)(&cfg.Controller),
		commandHandler(runtime),
	)
}

func commandHandler(r *rt.Runtime) ctrl.CommandFunc {
	return func(ctx context.Context, cmd *api.Command) error {
		switch msg := cmd.GetData().(type) {
		case *api.Command_Discovery:
			plugin := r.GetPluginsManager().Plugin(discovery.Name)
			if plugin == nil {
				return errors.Errorf("failed to find discovery plugin")
			}

			discoveryPlugin, ok := plugin.(*discovery.Discovery)
			if !ok {
				return errors.Errorf("failed to cast discovery plugin")
			}

			err := discoveryPlugin.Trigger(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to trigger discovery")
			}

		case *api.Command_SyncEdgeDirectory:
			plugin := r.GetPluginsManager().Plugin(edge.PluginName)
			if plugin == nil {
				return errors.Errorf("failed to find edge plugin")
			}

			edgePlugin, ok := plugin.(*edge.Plugin)
			if !ok {
				return errors.Errorf("failed to cast edge directory plugin")
			}

			edgePlugin.SyncNow(msg.SyncEdgeDirectory.GetMode())

		default:
			return errors.New("not implemented")
		}

		return nil
	}
}

func hostname() (string, error) {
	if host := os.Getenv(x.EnvAsertoHostName); host != "" {
		return host, nil
	}

	if host, err := os.Hostname(); err == nil && host != "" {
		return host, nil
	}

	if host := os.Getenv(x.EnvHostName); host != "" {
		return host, nil
	}

	return "", aerr.ErrBadRuntime.Msg("discovery hostname not set")
}

package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/runtime"
	"github.com/aserto-dev/topaz/pkg/config"
)

type OPAConfig runtime.Config

var _ config.Section = (*OPAConfig)(nil)

//nolint:mnd  // default values
func (c *OPAConfig) Defaults() map[string]any {
	return map[string]any{
		"instance_id":                      "-",
		"graceful_shutdown_period_seconds": 2,
		"max_plugin_wait_time_seconds":     30,
		"local_bundles.skip_verification":  true,
	}
}

func (c *OPAConfig) Validate() error {
	return nil
}

func (c *OPAConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("OPA").
		Funcs(config.TemplateFuncs()).
		Parse(config.TrimN(opaConfigTemplate))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func (c *OPAConfig) TryLocalBundles() *runtime.LocalBundlesConfig {
	if c.HasLocalBundles() {
		return &c.LocalBundles
	}

	return nil
}

func (c *OPAConfig) TryConfig() *runtime.OPAConfig {
	if c.HasConfig() {
		return &c.Config
	}

	return nil
}

func (c *OPAConfig) HasLocalBundles() bool {
	lb := &c.LocalBundles

	return lb.Watch || lb.SkipVerification || lb.LocalPolicyImage != "" || lb.FileStoreRoot != "" ||
		len(lb.Paths) > 0 || len(lb.Ignore) > 0 || lb.VerificationConfig != nil
}

func (c *OPAConfig) HasConfig() bool {
	cfg := &c.Config

	return len(cfg.Services) > 0 || len(cfg.Labels) > 0 || cfg.Discovery != nil || len(cfg.Bundles) > 0 ||
		cfg.DecisionLogs != nil || cfg.Status != nil || len(cfg.Plugins) > 0 || len(cfg.Keys) > 0 ||
		cfg.DefaultDecision != nil || cfg.DefaultAuthorizationDecision != nil || cfg.Caching != nil ||
		cfg.PersistenceDirectory != nil
}

const opaConfigTemplate = `
# Open Policy Agent configuration.
opa:
  instance_id: '{{ .InstanceID }}'
  graceful_shutdown_period_seconds: {{ .GracefulShutdownPeriodSeconds }}
  max_plugin_wait_time_seconds: {{ .MaxPluginWaitTimeSeconds }}
{{- with .TryLocalBundles }}
  local_bundles:
    paths: {{ .Paths }}

    {{- with .Ignore }}
    ignore: {{ . }}
    {{- end }}

    {{- with .LocalPolicyImage }}
    local_policy_image: {{ . }}
    {{- end }}

    {{- with .FileStoreRoot }}
    file_store_root: {{ .}}
    {{- end }}

    {{- with .Watch }}
    watch: {{ . }}
    {{- end }}
    skip_verification: {{ .SkipVerification }}
{{- end }}

{{- with .TryConfig }}
  config:
    {{- . | toMap | toYaml | nindent 4 }}
{{- end }}
`

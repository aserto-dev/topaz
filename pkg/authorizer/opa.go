package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/runtime"
)

type OPAConfig struct {
	runtime.Config
}

func (c *OPAConfig) Defaults() map[string]any {
	return map[string]any{}
}

func (c *OPAConfig) Validate() (bool, error) {
	return true, nil
}

func (c *OPAConfig) Generate(w io.Writer) error {
	tmpl, err := template.New("OPA").Parse(opaConfigTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const opaConfigTemplate = `
  # Open Policy Agent configuration.
  opa:
    instance_id: '{{ .InstanceID }}'
    graceful_shutdown_period_seconds: {{ .GracefulShutdownPeriodSeconds }}
    max_plugin_wait_time_seconds: {{ .MaxPluginWaitTimeSeconds }}
    {{- if .LocalBundles }}
    local_bundles:
      {{- if .LocalBundles.Paths }}
      paths: {{ .LocalBundles.Paths }}
      {{ end -}}
      {{- if .LocalBundles.Ignore }}
      ignore: {{ .LocalBundles.Ignore }}
      {{ end -}}
      {{- if .LocalBundles.LocalPolicyImage }}
      local_policy_image: {{ .LocalBundles.LocalPolicyImage}}
      {{ end -}}
      {{- if .LocalBundles.FileStoreRoot }}
      file_store_root: {{ .LocalBundles.FileStoreRoot}}
      {{ end -}}
      watch: {{ .LocalBundles.Watch }}
      skip_verification: {{ .LocalBundles.SkipVerification }}
    {{ end -}}
    config:
      services:
        ghcr:
          url: https://ghcr.io
          type: "oci"
          response_header_timeout_seconds: 5
      bundles:
        test:
          service: ghcr
          resource: "ghcr.io/aserto-policies/policy-rebac:latest"
          persist: false
          config:
            polling:
              min_delay_seconds: 60
              max_delay_seconds: 120
`

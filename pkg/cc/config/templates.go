package config

import "github.com/aserto-dev/topaz/pkg/cc/config/tpl"

type templateParams struct {
	Version          int
	LocalPolicyImage string
	PolicyName       string
	Resource         string
	Authorization    string
	EdgeDirectory    bool

	TenantID     string
	DiscoveryURL string
	TenantKey    string
	ControlPlane struct {
		Enabled        bool
		Address        string
		ClientCertPath string
		ClientKeyPath  string
	}
	DecisionLogging bool
	DecisionLogger  struct {
		EMSAddress     string
		StorePath      string
		ClientCertPath string
		ClientKeyPath  string
	}
}

var LocalImageTemplate = tpl.Preamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    local_policy_image: {{ .LocalPolicyImage }}
    watch: true
    skip_verification: true
`

var Template = tpl.Preamble + `
opa:
  instance_id: "-"
  graceful_shutdown_period_seconds: 2
  # max_plugin_wait_time_seconds: 30 set as default
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      ghcr:
        url: https://ghcr.io
        type: "oci"
        response_header_timeout_seconds: 5
    bundles:
      {{ .PolicyName }}:
        service: ghcr
        resource: "{{ .Resource }}"
        persist: false
        config:
          polling:
            min_delay_seconds: 60
            max_delay_seconds: 120
`

var EdgeTemplate = tpl.Preamble + `
opa:
  instance_id: {{ .TenantID }}
  graceful_shutdown_period_seconds: 2
  local_bundles:
    paths: []
    skip_verification: true
  config:
    services:
      aserto-discovery:
        url: {{ .DiscoveryURL }}
        credentials:
          bearer:
            token: {{ .TenantKey }}
            scheme: "basic"
        headers:
          Aserto-Tenant-Id: {{ .TenantID }}
    discovery:
      service: aserto-discovery
      resource: {{ .PolicyName }}/{{ .PolicyName }}/opa
{{ if .ControlPlane.Enabled }}
controller:
  enabled: true
  server:
    address: {{ .ControlPlane.Address }}
    client_cert_path: {{ .ControlPlane.ClientCertPath }}
    client_key_path: {{ .ControlPlane.ClientKeyPath }}
{{ else }}
controller:
  enabled: false
{{ end }}
{{ if .DecisionLogging }}
decision_logger:
  type: self
  config:
    store_directory: {{ .DecisionLogger.StorePath }}
    scribe:
      address: {{ .DecisionLogger.EMSAddress }}
      client_cert_path: {{ .DecisionLogger.ClientCertPath }}
      client_key_path: {{ .DecisionLogger.ClientKeyPath }}
      ack_wait_seconds: 30
      headers:
        Aserto-Tenant-Id: {{ .TenantID }}
    shipper:
      publish_timeout_seconds: 2
{{ end }}
`

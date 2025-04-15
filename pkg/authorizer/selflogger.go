package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/self-decision-logger/logger/self"
)

type SelfDecisionLoggerConfig self.Config

const SelfDecisionLoggerPlugin string = `self`

//nolint:mnd  // default values
func (c *SelfDecisionLoggerConfig) Defaults() map[string]any {
	return map[string]any{
		"port":                            4222,
		"store_directory":                 "./nats_store",
		"scribe.address":                  "ems.prod.aserto.com:8443",
		"scribe.client_cert_path":         "${TOPAZ_DIR}/certs/sidecar-prod.crt",
		"scribe.client_key_path":          "${TOPAZ_DIR}/certs/sidecar-prod.key",
		"scribe.ack_wait_seconds":         30,
		"shipper.max_bytes":               100 * 1024 * 1024, // 100MB
		"shipper.max_batch_size":          512,
		"shipper.publish_timeout_seconds": 10,
		"shipper.max_inflight_batches":    10,
		"shipper.ack_wait_seconds":        60,
		"shipper.backoff_seconds":         []int{5, 10, 30, 60, 120, 300},
	}
}

func (c *SelfDecisionLoggerConfig) Validate() (bool, error) {
	return true, nil
}

func (c *SelfDecisionLoggerConfig) Generate(w io.Writer) error {
	tmpl, err := template.New("SELF_DECISION_LOGGER").Parse(selfDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const selfDecisionLoggerTemplate string = `
port: {{ .Port }}
store_directory: '{{ .StoreDirectory }}'
scribe:
  address: '{{ .Scribe.Address }}'
  client_cert_path: '{{ .Scribe.ClientCertPath }}'
  client_key_path: '{{ .Scribe.ClientKeyPath }}'
  ack_wait_seconds: {{ .Scribe.AckWaitSeconds }}
  tenant_id: {{ .Scribe.TenantID }}
shipper:
  publish_timeout_seconds: 2
`

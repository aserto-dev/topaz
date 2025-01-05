package authorizer

import (
	"os"
	"text/template"

	"github.com/aserto-dev/self-decision-logger/logger/self"
	"github.com/aserto-dev/topaz/decision_log/logger/file"
	"github.com/aserto-dev/topaz/pkg/config/handler"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type DecisionLoggerConfig struct {
	Enabled  bool                   `json:"enabled"`
	Plugin   string                 `json:"plugin"`
	Settings map[string]interface{} `json:"settings"`
}

var _ = handler.Config(&DecisionLoggerConfig{})

func (c *DecisionLoggerConfig) SetDefaults(v *viper.Viper, p ...string) {
}

func (c *DecisionLoggerConfig) Validate() (bool, error) {
	return true, nil
}

func (c *DecisionLoggerConfig) Generate(w *os.File) error {
	if !c.Enabled {
		c.Plugin = DisabledDecisionLoggerPlugin
	}

	switch {
	case !c.Enabled:
		cfg := DisabledDecisionLoggerConfig{Enabled: c.Enabled}
		return cfg.Generate(w)
	case c.Plugin == FileDecisionLoggerPlugin:
		cfg := FileDecisionLoggerConfigFromMap(c.Settings)
		return cfg.Generate(w)
	case c.Plugin == SelfDecisionLoggerPlugin:
		cfg := SelfDecisionLoggerConfigFromMap(c.Settings)
		return cfg.Generate(w)
	default:
		return errors.Errorf("unknown store plugin %q", c.Plugin)
	}
}

type DisabledDecisionLoggerConfig struct {
	Enabled bool `json:"enabled"`
}

func (c *DisabledDecisionLoggerConfig) Generate(w *os.File) error {
	tmpl, err := template.New("DISABLED_DECISION_LOGGER").Parse(disabledDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const DisabledDecisionLoggerPlugin string = `disabled`

const disabledDecisionLoggerTemplate string = `
  # decision logger configuration.
  decision_logger:
    enabled: {{ .Enabled }}
`

type FileDecisionLoggerConfig struct {
	file.Config
}

func (c *FileDecisionLoggerConfig) Generate(w *os.File) error {
	tmpl, err := template.New("FILE_DECISION_LOGGER").Parse(fileDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func (c FileDecisionLoggerConfig) Map() map[string]interface{} {
	var m map[string]interface{}

	if err := mapstructure.Decode(c, &m); err != nil {
		return nil
	}

	return m
}

func FileDecisionLoggerConfigFromMap(m map[string]interface{}) *FileDecisionLoggerConfig {
	var cfg FileDecisionLoggerConfig
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

const FileDecisionLoggerPlugin string = `file`

const fileDecisionLoggerTemplate string = `
  # decision logger configuration.
  decision_logger:
		enabled: {{ .Enabled }}
    plugin: file
    settings:
      log_file_path: '{{ .LogFilePath }}'
      max_file_size_mb: {{ .MaxFileSizeMB }}
      max_file_count: {{ .MaxFileCount }}
`

type SelfDecisionLoggerConfig struct {
	self.Config
}

const SelfDecisionLoggerPlugin string = `self`

const selfDecisionLoggerTemplate string = `
  # decision logger configuration.
  decision_logger:
		enabled: {{ .Enabled }}
    plugin: self
    settings:
      store_directory: '{{ .StoreDirectory }}'
      scribe:
        address: ems.prod.aserto.com:8443
        client_cert_path: "${TOPAZ_DIR}/certs/sidecar-prod.crt"
        client_key_path: "${TOPAZ_DIR}/certs/sidecar-prod.key"
        ack_wait_seconds: 30
        headers:
          Aserto-Tenant-Id: 55cf8ea9-30b2-4f9a-b0bb-021ca12170f3
      shipper:
        publish_timeout_seconds: 2
`

func (c *SelfDecisionLoggerConfig) Generate(w *os.File) error {
	tmpl, err := template.New("SELF_DECISION_LOGGER").Parse(selfDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

func (c SelfDecisionLoggerConfig) Map() map[string]interface{} {
	var m map[string]interface{}

	if err := mapstructure.Decode(c, &m); err != nil {
		return nil
	}

	return m
}

func SelfDecisionLoggerConfigFromMap(m map[string]interface{}) *SelfDecisionLoggerConfig {
	var cfg SelfDecisionLoggerConfig
	if err := mapstructure.Decode(m, &cfg); err != nil {
		return nil
	}

	return &cfg
}

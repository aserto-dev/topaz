package authorizer

import (
	"io"
	"iter"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/authorizer/decisionlog/logger/file"
	"github.com/aserto-dev/topaz/topazd/loiter"
	"github.com/samber/lo"
)

type FileDecisionLoggerConfig file.Config

const FileDecisionLoggerPlugin string = `file`

var _ config.Section = (*FileDecisionLoggerConfig)(nil)

//nolint:mnd  // default values
func (c *FileDecisionLoggerConfig) Defaults() map[string]any {
	return map[string]any{
		"log_file_path":    ".",
		"max_file_size_mb": 50,
		"max_file_count":   2,
	}
}

func (c *FileDecisionLoggerConfig) Validate() error {
	return nil
}

func (c *FileDecisionLoggerConfig) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Seq2(lo.T2(c.LogFilePath, config.ReadWrite))
}

func (c *FileDecisionLoggerConfig) Serialize(w io.Writer) error {
	tmpl, err := template.New("FILE_DECISION_LOGGER").Parse(fileDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	type params struct {
		*FileDecisionLoggerConfig

		Provider_ string
	}

	p := &params{c, FileDecisionLoggerPlugin}
	if err := tmpl.Execute(w, p); err != nil {
		return err
	}

	return nil
}

const fileDecisionLoggerTemplate string = `
{{ .Provider_ }}:
  log_file_path: '{{ .LogFilePath }}'
  max_file_size_mb: {{ .MaxFileSizeMB }}
  max_file_count: {{ .MaxFileCount }}
`

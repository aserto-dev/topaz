package authorizer

import (
	"io"
	"text/template"

	"github.com/aserto-dev/topaz/decisionlog/logger/file"
)

type FileDecisionLoggerConfig file.Config

const FileDecisionLoggerPlugin string = `file`

//nolint:mnd  // default values
func (c *FileDecisionLoggerConfig) Defaults() map[string]any {
	return map[string]any{
		"log_file_path":    ".",
		"max_file_size_mb": 50,
		"max_file_count":   2,
	}
}

func (c *FileDecisionLoggerConfig) Validate() (bool, error) {
	return true, nil
}

func (c *FileDecisionLoggerConfig) Generate(w io.Writer) error {
	tmpl, err := template.New("FILE_DECISION_LOGGER").Parse(fileDecisionLoggerTemplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	return nil
}

const fileDecisionLoggerTemplate string = `
file:
  log_file_path: '{{ .LogFilePath }}'
  max_file_size_mb: {{ .MaxFileSizeMB }}
  max_file_count: {{ .MaxFileCount }}
`

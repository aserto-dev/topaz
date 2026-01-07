package authorizer

import (
	"fmt"
	"io"
	"iter"
	"text/template"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topazd/loiter"
	"github.com/samber/lo"
)

type Config struct {
	OPA            OPAConfig            `json:"opa"`
	DecisionLogger DecisionLoggerConfig `json:"decision_logger"`
	Controller     ControllerConfig     `json:"controller"`
	JWT            JWTConfig            `json:"jwt"`
}

var _ config.Section = (*Config)(nil)

func (c *Config) Defaults() map[string]any {
	return lo.Assign(
		config.PrefixKeys("opa", c.OPA.Defaults()),
		config.PrefixKeys("decision_logger", c.DecisionLogger.Defaults()),
		config.PrefixKeys("controller", c.Controller.Defaults()),
		config.PrefixKeys("jwt", c.JWT.Defaults()),
	)
}

func (*Config) Validate() error {
	return nil
}

func (c *Config) Paths() iter.Seq2[string, config.AccessMode] {
	return loiter.Chain2(
		c.OPA.Paths(),
		c.DecisionLogger.Paths(),
		c.Controller.Paths(),
		c.JWT.Paths(),
	)
}

func (c *Config) Serialize(w io.Writer) error {
	tmpl, err := template.New("AUTHORIZER").Parse(config.TrimN(configTemplate))
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, c); err != nil {
		return err
	}

	wi := config.IndentWriter(w, sectionIndentLevel)

	if err := c.OPA.Serialize(wi); err != nil {
		return err
	}

	// Write newlines between sections.
	// If the newline would be part of the section template it would be indented and result in an
	// all-spaces line (i.e. "  \n").
	_, _ = fmt.Fprint(w, "\n")

	if err := c.DecisionLogger.Serialize(wi); err != nil {
		return err
	}

	_, _ = fmt.Fprint(w, "\n")

	if err := c.Controller.Serialize(wi); err != nil {
		return err
	}

	_, _ = fmt.Fprint(w, "\n")

	return c.JWT.Serialize(wi)
}

const (
	sectionIndentLevel = 2

	configTemplate = `
# authorizer configuration.
authorizer:
`
)

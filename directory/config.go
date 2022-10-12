package directory

import (
	"github.com/aserto-dev/edge-ds/pkg/directory"
	eds "github.com/aserto-dev/go-eds"
)

type Config struct {
	Path       string           `json:"path"` // backwards compatibility to create eds.Config
	EdgeConfig directory.Config `json:"edge"`
	Remote     struct {
		Addr     string `json:"address"`
		Key      string `json:"api_key"`
		Insecure bool   `json:"insecure"`
		TenantID string `json:"tenant_id"`
	} `json:"remote"`
}

func (c *Config) EDSPath() *eds.Config {
	return &eds.Config{Path: c.Path}
}

package directory

import (
	eds "github.com/aserto-dev/go-eds"
)

type Config struct {
	Path   string `json:"path"` // backwards compatibility to create eds.Config
	Remote struct {
		Addr     string `json:"address"`
		Key      string `json:"api_key"`
		Insecure bool   `json:"insecure"`
		TenantID string `json:"tenant_id"`
	} `json:"remote"`
	IsHosted bool `json:"is_hosted"`
}

func (c *Config) EDSPath() *eds.Config {
	return &eds.Config{Path: c.Path}
}

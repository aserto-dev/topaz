package directory

import (
	"github.com/aserto-dev/edge-ds/pkg/directory"
)

type Config struct {
	EdgeConfig directory.Config `json:"edge"`
	Remote     struct {
		Addr     string `json:"address"`
		Key      string `json:"api_key"`
		Insecure bool   `json:"insecure"`
		TenantID string `json:"tenant_id"`
	} `json:"remote"`
}

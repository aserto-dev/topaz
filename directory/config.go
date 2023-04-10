package directory

import "time"

type Config struct {
	EdgeConfig struct {
		DBPath         string        `json:"db_path"`
		RequestTimeout time.Duration `json:"request_timeout"`
		Seed           bool          `json:"seed_metadata"`
	} `json:"edge"`
	Remote struct {
		Addr     string `json:"address"`
		Key      string `json:"api_key"`
		Insecure bool   `json:"insecure"`
		TenantID string `json:"tenant_id"`
	} `json:"remote"`
}

package authzen

type Config struct {
	Enabled bool `json:"enabled"`
}

func (c Config) IsEnabled() bool {
	return c.Enabled
}

package config

type currentConfig struct {
	*Loader
	err error
}

func GetConfig(configFilePath string) *currentConfig {
	cfg, err := LoadConfiguration(configFilePath)
	if err != nil {
		return &currentConfig{Loader: nil, err: err}
	}

	return &currentConfig{Loader: cfg, err: nil}
}

func (c *currentConfig) Ports() ([]string, error) {
	if c.err != nil {
		return []string{}, c.err
	}

	return c.GetPorts()
}

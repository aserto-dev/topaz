package common_test

import "time"

const (
	TestStartupTimeout        = 300 * time.Second
	TestConfigFilePath string = "/config/config.yaml"
	TestDBFilePath     string = "/data/test.db"
)

var TestExposedPorts = []string{"9292/tcp"}

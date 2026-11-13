package topaz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aserto-dev/topaz/pkg/config"
	"github.com/aserto-dev/topaz/topaz/cc"
	"github.com/cli/browser"
)

const (
	consoleService     = "console"
	consoleInitialPath = "/ui/directory/model"
)

var openConsole = browser.OpenURL

type ConsoleCmd struct {
	ConsoleAddress string `arg:"" optional:"" help:"gateway address of the console service"`
}

func (cmd *ConsoleCmd) Run(_ context.Context) error {
	consoleAddress := cmd.ConsoleAddress
	if consoleAddress == "" {
		cliCfg, err := cc.RequireConfig()
		if err != nil {
			return err
		}

		topazCfg, err := getSvcConfig(cliCfg)
		if err != nil {
			return err
		}

		consoleAddress, err = configuredConsoleAddress(topazCfg)
		if err != nil {
			return err
		}
	}

	return openConsole(consoleURL(consoleAddress))
}

func getSvcConfig(cliCfg *cc.Config) (*config.Config, error) {
	configFile := cliCfg.Running.ConfigFile
	if configFile == "" {
		configFile = cliCfg.Active.ConfigFile
	}
	if configFile == "" {
		return nil, cc.ErrNoConfig
	}

	loader, err := config.LoadConfiguration(configFile)
	if err != nil {
		return nil, fmt.Errorf("load Topaz configuration %q: %w", configFile, err)
	}

	return loader.Configuration, nil
}

func configuredConsoleAddress(cfg *config.Config) (string, error) {
	service, ok := cfg.APIConfig.Services[consoleService]
	if !ok || service == nil {
		return "", errors.New("console service is not configured")
	}

	address := service.Gateway.FQDN
	if address == "" {
		address = service.Gateway.ListenAddress
	}
	if address == "" {
		return "", errors.New("console gateway address is not configured")
	}

	address = clientAddress(address)
	if !strings.Contains(address, "://") {
		scheme := "https://"
		if !service.Gateway.Certs.HasCert() {
			scheme = "http://"
		}

		address = scheme + address
	}

	return address, nil
}

func consoleURL(address string) string {
	address = strings.TrimRight(strings.TrimSpace(address), "/")
	if !strings.Contains(address, "://") {
		address = "https://" + address
	}
	if !strings.Contains(address, "/ui/") {
		address += consoleInitialPath
	} else if strings.HasSuffix(address, "/ui/directory") {
		address += "/model"
	}

	return address
}

func clientAddress(address string) string {
	return strings.Replace(address, "0.0.0.0", "localhost", 1)
}

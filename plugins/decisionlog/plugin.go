package decisionlog

import (
	"context"

	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/topaz/decisionlog"
	"github.com/aserto-dev/topaz/pkg/config"

	"github.com/open-policy-agent/opa/v1/plugins"
)

const PluginName = "aserto_decision_log"

type PolicyInfo struct {
	PolicyID        string `json:"policy_id"`
	PolicyName      string `json:"policy_name"`
	InstanceLabel   string `json:"instance_label"` // DO NOT REMOVE InstanceLabel, required by discovery.
	RegistryService string `json:"registry_service"`
	RegistryImage   string `json:"registry_image"`
	RegistryTag     string `json:"registry_tag"`
	Digest          string `json:"digest"`
}
type Config struct {
	config.Optional

	PolicyInfo PolicyInfo `json:"policy_info"`
}
type DecisionLogsPlugin struct {
	manager *plugins.Manager
	cfg     *Config
	logger  decisionlog.DecisionLogger
}

func newDecisionLogger(cfg *Config, manager *plugins.Manager, logger decisionlog.DecisionLogger) *DecisionLogsPlugin {
	return &DecisionLogsPlugin{
		manager: manager,
		cfg:     cfg,
		logger:  logger,
	}
}

func (plugin *DecisionLogsPlugin) Start(ctx context.Context) error {
	plugin.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})
	return nil
}

func (plugin *DecisionLogsPlugin) Stop(ctx context.Context) {
	plugin.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}

func (plugin *DecisionLogsPlugin) Reconfigure(ctx context.Context, config any) {
	plugin.cfg, _ = config.(*Config)
}

func (plugin *DecisionLogsPlugin) Log(ctx context.Context, d *api.Decision) error {
	if !plugin.cfg.Enabled || plugin.logger == nil {
		return nil
	}

	d.Policy.RegistryService = plugin.cfg.PolicyInfo.RegistryService
	d.Policy.RegistryImage = plugin.cfg.PolicyInfo.RegistryImage
	d.Policy.RegistryTag = plugin.cfg.PolicyInfo.RegistryTag
	d.Policy.RegistryDigest = plugin.cfg.PolicyInfo.Digest

	return plugin.logger.Log(d)
}

func Lookup(m *plugins.Manager) *DecisionLogsPlugin {
	p := m.Plugin(PluginName)
	if p == nil {
		return nil
	}

	plugin, _ := p.(*DecisionLogsPlugin)

	return plugin
}

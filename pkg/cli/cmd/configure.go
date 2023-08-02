package cmd

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"strings"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-grpc/aserto/tenant/connection/v1"
	"github.com/aserto-dev/header"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/clients"
	"github.com/fatih/color"
	"google.golang.org/protobuf/types/known/structpb"
)

type ConfigureCmd struct {
	PolicyName       string `short:"n" help:"policy name"`
	LocalPolicyImage string `short:"l" help:"local policy image name"`
	Resource         string `short:"r" help:"resource url"`
	Stdout           bool   `short:"p" help:"generated configuration is printed to stdout but not saved"`
	EdgeDirectory    bool   `short:"d" help:"enable edge directory" default:"false"`
	SeedMetadata     bool   `short:"s" help:"enable seed metadata" default:"false"`

	EdgeAuthorizer    bool   `short:"e" help:"configure topaz to work as an edge authorizer connected to the aserto control plane" default:"false"`
	TenantAddress     string `help:"aserto tenant service address" default:"tenant.prod.aserto.com:8443"`
	TenantID          string `help:"your aserto tenant id"`
	TenantAPIKey      string `help:"root API key to connect to the tenant service"`
	TenantToken       string `help:"token to connect to the tenant service"`
	ConnectionID      string `help:"edge authorizer connection id"`
	LogStoreDirectory string `help:"local path to store decision logs" default:"decision-logs"`
}

func (cmd *ConfigureCmd) Run(c *cc.CommonCtx) error {
	if cmd.PolicyName == "" && cmd.Resource == "" {
		if cmd.LocalPolicyImage == "" {
			return errors.New("you either need to provide a local policy image or the resource and the policy name for the configuration")
		}
	}
	color.Green(">>> configure policy")

	configDir, err := CreateConfigDir()
	if err != nil {
		return err
	}

	if _, err := CreateCertsDir(); err != nil {
		return err
	}

	if _, err := CreateDataDir(); err != nil {
		return err
	}

	params := templateParams{
		LocalPolicyImage: cmd.LocalPolicyImage,
		PolicyName:       cmd.PolicyName,
		Resource:         cmd.Resource,
		EdgeDirectory:    cmd.EdgeDirectory,
		SeedMetadata:     cmd.SeedMetadata,
	}

	if cmd.EdgeAuthorizer && cmd.TenantAddress != "" {

		clientConfig := clients.TenantConfig{
			Address:  cmd.TenantAddress,
			Token:    cmd.TenantToken,
			APIKey:   cmd.TenantAPIKey,
			TenantID: cmd.TenantID,
			Insecure: true,
		}

		client, err := clients.NewTenantConnectionClient(c, &clientConfig)
		if err != nil {
			return err
		}

		tenantIDContext := header.ContextWithTenantID(c.Context, cmd.TenantID)

		certFile, keyFile, err := getEdgeAuthorizerCerts(tenantIDContext, client, cmd.ConnectionID, configDir)
		if err != nil {
			return err
		}
		params.EdgeAuthorzier = cmd.EdgeAuthorizer
		params.EdgeCertFile = certFile
		params.EdgeKeyFile = keyFile
		tenantArr := strings.Split(cmd.TenantAddress, ".")
		tenantArr[0] = "ems"
		params.EMSAddress = strings.Join(tenantArr, ".")
		tenantArr[0] = "relay"
		params.RelayAddress = strings.Join(tenantArr, ".")
		params.TenantID = cmd.TenantID

		discoveryConfig, err := getDiscoveryConfig(c.Context, client)
		if err != nil {
			return err
		}
		params.DiscoveryKey = discoveryConfig.APIKey
		params.DiscoveryURL = discoveryConfig.URL

		params.LogStoreDirectory = cmd.LogStoreDirectory
	}

	var w io.Writer

	if cmd.Stdout {
		w = c.UI.Output()
	} else {
		w, err = os.Create(path.Join(configDir, "config.yaml"))
		if err != nil {
			return err
		}
	}

	if params.LocalPolicyImage != "" {
		color.Green("using local policy image: %s", params.LocalPolicyImage)
		return WriteConfig(w, localImageTemplate, &params)
	}

	color.Green("policy name: %s", params.PolicyName)

	return WriteConfig(w, configTemplate, &params)
}

func CreateConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := path.Join(home, "/.config/topaz/cfg")
	if fi, err := os.Stat(configDir); err == nil && fi.IsDir() {
		return configDir, nil
	}

	return configDir, os.MkdirAll(configDir, 0700)
}

func CreateCertsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	certsDir := path.Join(home, "/.config/topaz/certs")
	if fi, err := os.Stat(certsDir); err == nil && fi.IsDir() {
		return certsDir, nil
	}

	return certsDir, os.MkdirAll(certsDir, 0700)
}

func CreateDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dataDir := path.Join(home, "/.config/topaz/db")
	if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
		return dataDir, nil
	}

	return dataDir, os.MkdirAll(dataDir, 0700)
}

func WriteConfig(w io.Writer, templ string, params *templateParams) error {
	t, err := template.New("config").Parse(templ)
	if err != nil {
		return err
	}

	err = t.Execute(w, params)
	if err != nil {
		return err
	}

	return nil
}

func getEdgeAuthorizerCerts(ctx context.Context, client connection.ConnectionClient, connectionID, configDir string) (certFile, keyFile string, err error) {
	resp, err := client.GetConnection(ctx, &connection.GetConnectionRequest{
		Id: connectionID,
	})
	if err != nil {
		return "", "", err
	}

	conn := resp.Result
	if conn == nil {
		return "", "", errors.New("invalid empty connection")
	}

	if conn.Kind != api.ProviderKind_PROVIDER_KIND_EDGE_AUTHORIZER {
		return "", "", errors.New("not an edge authorizer connection")
	}

	certs := conn.Config.Fields["api_cert"].GetListValue().GetValues()
	if len(certs) == 0 {
		return "", "", errors.New("invalid configuration: api_cert")
	}

	structVal := certs[len(certs)-1].GetStructValue()
	if structVal == nil {
		return "", "", errors.New("invalid configuration: api_cert")
	}

	err = fileFromConfigField(structVal, "certificate", configDir, "client.crt")
	if err != nil {
		return "", "", err
	}

	err = fileFromConfigField(structVal, "private_key", configDir, "client.key")
	if err != nil {
		return "", "", err
	}

	return "client.crt", "client.key", nil
}

func fileFromConfigField(structVal *structpb.Struct, field, configDir, fileName string) error {
	val, ok := structVal.Fields[field]
	if !ok {
		return fmt.Errorf("missing field: %s", field)
	}

	strVal := val.GetStringValue()
	if strVal == "" {
		return fmt.Errorf("empty field: %s", field)
	}

	filePath := path.Join(configDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strVal)
	if err != nil {
		return err
	}

	return nil
}

type discoveryConfig struct {
	URL    string
	APIKey string
}

func newDiscoveryConfig(config *structpb.Struct) (*discoveryConfig, error) {
	urlField, ok := config.Fields["url"]
	if !ok {
		return nil, errors.New("missing field: url")
	}

	apiKeyField, ok := config.Fields["api_key"]
	if !ok {
		return nil, errors.New("missing field: api_key")
	}

	return &discoveryConfig{URL: urlField.GetStringValue(), APIKey: apiKeyField.GetStringValue()}, nil
}

func getDiscoveryConfig(ctx context.Context, client connection.ConnectionClient) (*discoveryConfig, error) {
	resp, err := client.ListConnections(
		ctx,
		&connection.ListConnectionsRequest{Kind: api.ProviderKind_PROVIDER_KIND_DISCOVERY},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, errors.New("no discovery connections available for tenant. please contact support@aserto.com")
	}

	for _, conn := range resp.Results {
		conResp, err := client.GetConnection(ctx, &connection.GetConnectionRequest{Id: conn.Id})
		if err == nil {
			conf, err := newDiscoveryConfig(conResp.Result.Config)
			if err == nil {
				return conf, nil
			}
		}
	}

	return nil, errors.New("cannot find discovery configuration")
}

package clients

type Config struct {
	Host     string `flag:"host" short:"H" env:"HOST" help:"directory service address"`
	APIKey   string `flag:"api-key" short:"k" env:"API_KEY" help:"directory API key"`
	Token    string `flag:"token" short:"t" env:"TOKEN" help:"JWT used for connection"`
	Insecure bool   `flag:"insecure" short:"i" env:"INSECURE" help:"skip TLS verification"`
	TenantID string `flag:"tenant-id" help:"" env:"TENANT_ID" `
}

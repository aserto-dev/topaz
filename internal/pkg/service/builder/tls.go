package builder

import "github.com/aserto-dev/go-aserto"

func NoTLS(cfg *aserto.TLSConfig) bool {
	return (cfg == nil || cfg.CA == "" || cfg.Key == "" || cfg.Cert == "")
}

func TLS(cfg *aserto.TLSConfig) bool {
	return !NoTLS(cfg)
}

package config

import (
	"net"

	"github.com/aserto-dev/go-aserto"
)

type OptionalListener interface {
	IsEnabled() bool
	Port() string
}

// Optional is a configuration mixin for features that can be enabled or disabled.
type Optional struct {
	Enabled bool `json:"enabled"`
}

func (o *Optional) IsEnabled() bool {
	return o.Enabled
}

// Listener is a configuration mixin for sections that represent a socket listener.
type Listener struct {
	ListenAddress string           `json:"listen_address"`
	Certs         aserto.TLSConfig `json:"certs"`
}

func (l *Listener) IsEmptyAddress() bool {
	return l == nil || l.ListenAddress == ""
}

func (l *Listener) Port() string {
	_, port, _ := net.SplitHostPort(l.ListenAddress)
	return port
}

func (l *Listener) TryCerts() *aserto.TLSConfig {
	zero := aserto.TLSConfig{}
	if l.Certs == zero {
		return nil
	}

	return &l.Certs
}

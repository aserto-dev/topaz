package config

import (
	"iter"
	"net"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/topaz/pkg/loiter"
)

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

func (l *Listener) Paths() iter.Seq2[string, AccessMode] {
	return loiter.WithValue(
		loiter.Seq(l.Certs.Key, l.Certs.Cert, l.Certs.CA),
		ReadOnly,
	)
}

type OptionalListener interface {
	IsEnabled() bool
	Port() string
}

func ClientCertPaths(c *aserto.Config) iter.Seq2[string, AccessMode] {
	return loiter.WithValue(
		loiter.Seq(c.ClientKeyPath, c.ClientCertPath, c.CACertPath),
		ReadOnly,
	)
}

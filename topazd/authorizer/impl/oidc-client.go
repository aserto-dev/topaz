package impl

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	wellKnownOpenIDConfiguration string = `/.well-known/openid-configuration`
	fetchOidcConfigReqTimeout           = 5 * time.Second
	acceptHeader                 string = "Accept"
	contentTypeJSON              string = "application/json"
	dialerTimeout                       = 5 * time.Second
	dialerKeepAlive                     = 30 * time.Second
	transportMaxIdleConns               = 100
	transportIdleConnTimeout            = 90 * time.Second
	transportTLSHandshakeTimeout        = 10 * time.Second
	oidcClientTimeout                   = 10 * time.Second
)

type OidcClient struct {
	httpClient *http.Client
}

type OidcConfiguration struct {
	Issuer                string   `json:"issuer"`
	JwksURI               string   `json:"jwks_uri"`
	AuthorizationEndpoint string   `json:"authorization_endpoint,omitempty"`
	TokenEndpoint         string   `json:"token_endpoint,omitempty"`
	Algs                  []string `json:"id_token_signing_alg_values_supported,omitempty"`
}

func NewOidcClient() *OidcClient {
	dialer := &net.Dialer{
		Timeout:   dialerTimeout,
		KeepAlive: dialerKeepAlive,
	}

	transport := &http.Transport{
		// DialTLSContext intercepts BOTH network routing and the TLS handshake.
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid address format: %w", err)
			}

			// 1. Context-aware DNS resolution.
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("dns lookup failed: %w", err)
			}

			var targetIP net.IP

			for _, ipAddr := range ips {
				ip := ipAddr.IP
				// Strict SSRF Filter Block.
				if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
					continue // Skip unsafe IPs.
				}

				targetIP = ip

				break // Bind to the first valid public IP.
			}

			if targetIP == nil {
				return nil, errors.Errorf("no valid public IP addresses found for host %s", host)
			}

			// 2. DNS Rebinding, connect to the validated IP address directly.
			targetAddr := net.JoinHostPort(targetIP.String(), port)

			rawConn, err := dialer.DialContext(ctx, network, targetAddr)
			if err != nil {
				return nil, fmt.Errorf("network dial failed: %w", err)
			}

			// 3. Dynamic SNI Injection, build unique TLS config for connection.
			tlsConfig := &tls.Config{
				ServerName: host, // Inject the original hostname into the SNI extension.
				MinVersion: tls.VersionTLS12,
			}

			// Manually initiate secure TLS handshake using pinned IP.
			tlsConn := tls.Client(rawConn, tlsConfig)
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				rawConn.Close()
				return nil, fmt.Errorf("tls handshake failed: %w", err)
			}

			return tlsConn, nil
		},

		ForceAttemptHTTP2:   true,
		MaxIdleConns:        transportMaxIdleConns,
		IdleConnTimeout:     transportIdleConnTimeout,
		TLSHandshakeTimeout: transportTLSHandshakeTimeout,
	}

	return &OidcClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   oidcClientTimeout,
		},
	}
}

func (c *OidcClient) FetchAndValidateConfig(ctx context.Context, targetIssuer string) (*OidcConfiguration, error) {
	parsedTarget, err := url.Parse(targetIssuer)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL syntax: %w", err)
	}

	if parsedTarget.Scheme != "https" {
		return nil, errors.Errorf("insecure scheme %q: issuer must use https", parsedTarget.Scheme)
	}

	wellKnownURL := strings.TrimSuffix(targetIssuer, "/") + wellKnownOpenIDConfiguration

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(acceptHeader, contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var config OidcConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode json configuration: %w", err)
	}

	if config.Issuer != targetIssuer {
		return nil, errors.Errorf("issuer mismatch: expected %q, got %q", targetIssuer, config.Issuer)
	}

	return &config, nil
}

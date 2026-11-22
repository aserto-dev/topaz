package client_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/stretchr/testify/require"
)

func TestWithAddr(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithAddr("address"))
	assert.NoError(err)

	assert.Equal("address", options.Address)
}

func TestWithURL(t *testing.T) {
	assert := require.New(t)

	const URL = "https://server.com:123"

	svcURL, err := url.Parse(URL)
	assert.NoError(err)

	options, err := client.NewConnectionOptions(client.WithURL(svcURL))
	assert.NoError(err)

	assert.Equal(URL, options.Address)
}

func TestAddrAndURL(t *testing.T) {
	assert := require.New(t)
	svcURL, err := url.Parse("https://server.com:123")
	assert.NoError(err)

	_, err = client.NewConnectionOptions(client.WithAddr("address"), client.WithURL(svcURL))
	assert.Error(err)
}

func TestWithInsecure(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithInsecure(true))
	assert.NoError(err)

	assert.True(options.Insecure)
}

func TestWithTokenAuth(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithTokenAuth("<token>"))
	assert.NoError(err)

	md, err := options.Creds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("bearer <token>", token)
}

func TestWithBearerTokenAuth(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithTokenAuth("bearer <token>"))
	assert.NoError(err)

	md, err := options.Creds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("bearer <token>", token)
}

func TestWithAPIKey(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithAPIKeyAuth("<apikey>"))
	assert.NoError(err)

	md, err := options.Creds.GetRequestMetadata(context.TODO())
	assert.NoError(err)

	token, ok := md["authorization"]
	assert.True(ok)
	assert.Equal("basic <apikey>", token)
}

func TestTokenAndAPIKey(t *testing.T) {
	_, err := client.NewConnectionOptions(client.WithAPIKeyAuth("<apikey>"), client.WithTokenAuth("<token>"))
	require.Error(t, err)
}

const (
	caPath   = "/path/to/ca.crt"
	certPath = "/path/to/cert.crt"
	keyPath  = "/path/to/cert.key"
)

func TestWithCACertPath(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithCACertPath(caPath))
	assert.NoError(err)

	assert.Equal(caPath, options.CACertPath)
}

func TestWithClientCert(t *testing.T) {
	assert := require.New(t)

	options, err := client.NewConnectionOptions(client.WithClientCert(certPath, keyPath))
	assert.NoError(err)

	assert.Equal(certPath, options.ClientCertPath)
	assert.Equal(keyPath, options.ClientKeyPath)
}

func TestWithMissingClientCert(t *testing.T) {
	assert := require.New(t)

	certPath, keyPath := "", "/path/to/cert.key"
	_, err := client.NewConnectionOptions(client.WithClientCert(certPath, keyPath))
	assert.Error(err)
}

func TestWithMissingClientKey(t *testing.T) {
	assert := require.New(t)

	certPath, keyPath := "/path/to/cert.crt", ""
	_, err := client.NewConnectionOptions(client.WithClientCert(certPath, keyPath))
	assert.Error(err)
}

func TestWithHeader(t *testing.T) {
	assert := require.New(t)
	h1, v1 := "header1", "value1"
	h2, v2 := "header2", "value2"

	options, err := client.NewConnectionOptions(client.WithHeader(h1, v1), client.WithHeader(h2, v2))
	assert.NoError(err)

	a1, ok := options.Headers[h1]
	assert.True(ok)
	assert.Equal(v1, a1)

	a2, ok := options.Headers[h2]
	assert.True(ok)
	assert.Equal(v2, a2)
}

func TestWithNoTLS(t *testing.T) {
	assert := require.New(t)
	options, err := client.NewConnectionOptions(client.WithNoTLS(true))
	assert.NoError(err)
	assert.True(options.NoTLS)
}

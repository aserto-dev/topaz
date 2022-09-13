package testing

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aserto-dev/aserto-grpc/grpcclient"
	"github.com/aserto-dev/aserto-grpc/grpcutil"
	"github.com/aserto-dev/go-lib/grpc-clients/authorizer"
)

// CreateClient creates a new http client that can talk to the API
func (h *EngineHarness) CreateClient() *http.Client {
	caCert, err := os.ReadFile(h.Engine.Configuration.API.Gateway.Certs.TLSCACertPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS12,
			}}}

	return httpClient
}

func (h *EngineHarness) Req(verb, path, tenantID, body string) (string, int) {
	httpClient := h.CreateClient()
	url := "https://127.0.0.1:8383" + path
	req, err := http.NewRequestWithContext(context.Background(), verb, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		h.t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: use an API key

	req.Header.Set(string(grpcutil.HeaderAsertoTenantID), tenantID)

	resp, err := httpClient.Do(req)
	if err != nil {
		h.t.Error(err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.t.Error(err)
	}

	return string(responseBody), resp.StatusCode
}

func (h *EngineHarness) CreateGRPCClient() *authorizer.Client {
	grpcClient, err := authorizer.NewAuthorizerClient(
		h.Engine.Context,
		h.Engine.Logger,
		grpcclient.NewDialOptionsProvider(),
		&authorizer.ClientConfig{
			Config: grpcclient.Config{
				Address:    "127.0.0.1:8282",
				CACertPath: h.Engine.Configuration.API.GRPC.Certs.TLSCACertPath,
				// TODO: use an API key
				// https://github.com/aserto-dev/aserto-authorizer/blob/abd6625aacdea08e65a7796f03deb79c07486517/pkg/testing/client.go
			},
		},
	)

	if err != nil {
		h.t.Fatal(err)
	}

	return grpcClient
}

func (h *EngineHarness) CreateGRPCDirectoryClient() *authorizer.DirectoryClient {
	grpcClient, err := authorizer.NewDirectoryClient(
		h.Engine.Context,
		h.Engine.Logger,
		grpcclient.NewDialOptionsProvider(),
		&authorizer.ClientConfig{
			Config: grpcclient.Config{
				Address:    "127.0.0.1:8282",
				CACertPath: h.Engine.Configuration.API.GRPC.Certs.TLSCACertPath,
				// TODO: use an API key
				// https://github.com/aserto-dev/aserto-authorizer/blob/abd6625aacdea08e65a7796f03deb79c07486517/pkg/testing/client.go
			},
		},
	)

	if err != nil {
		h.t.Fatal(err)
	}

	return grpcClient
}

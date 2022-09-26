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
	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-lib/grpc-clients/authorizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func (h *EngineHarness) CreateGRPCClientV2() authz2.AuthorizerClient {
	var opts []grpc.DialOption
	var tlsConf tls.Config
	certPool := x509.NewCertPool()
	caCertBytes, err := os.ReadFile(h.Engine.Configuration.API.GRPC.Certs.TLSCACertPath)
	if err != nil {
		h.t.Fatal(err)
	}

	if !certPool.AppendCertsFromPEM(caCertBytes) {
		h.t.Fatal(err)
	}
	tlsConf.RootCAs = certPool
	tlsConf.MinVersion = tls.VersionTLS12

	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConf)))

	conn, err := grpc.Dial("127.0.0.1:8282", opts...)
	if err != nil {
		h.t.Fatal(err)
	}
	return authz2.NewAuthorizerClient(conn)
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

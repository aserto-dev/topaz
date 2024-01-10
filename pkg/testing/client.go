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

	"github.com/aserto-dev/go-aserto/client"
	authz2 "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	dse3 "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dsm3 "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	dsr3 "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dsw3 "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// CreateClient creates a new http client that can talk to the API.
func (h *EngineHarness) CreateClient() *http.Client {
	authorizerAPIConfig, ok := h.Engine.Configuration.APIConfig.Services["authorizer"]
	if !ok {
		log.Fatal("no authorizer configuration found")
	}
	caCert, err := os.ReadFile(authorizerAPIConfig.Gateway.Certs.TLSCACertPath)
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

func (h *EngineHarness) CreateGRPCClient() authz2.AuthorizerClient {

	authorizerAPIConfig, ok := h.Engine.Configuration.APIConfig.Services["authorizer"]
	if !ok {
		log.Fatal("no authorizer configuration found")
	}
	var opts []grpc.DialOption
	var tlsConf tls.Config
	certPool := x509.NewCertPool()
	caCertBytes, err := os.ReadFile(authorizerAPIConfig.GRPC.Certs.TLSCACertPath)
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

type DirectoryClient struct {
	Model    dsm3.ModelClient
	Reader   dsr3.ReaderClient
	Writer   dsw3.WriterClient
	Importer dsi3.ImporterClient
	Exporter dse3.ExporterClient
}

func (h *EngineHarness) CreateDirectoryClient(ctx context.Context) *DirectoryClient {
	readerAPIConfig, ok := h.Engine.Configuration.APIConfig.Services["reader"]
	if !ok {
		log.Fatal("no reader configuration found")
	}

	c, err := client.NewConnection(
		ctx,
		client.WithAddr(readerAPIConfig.GRPC.ListenAddress),
		client.WithCACertPath(readerAPIConfig.GRPC.Certs.TLSCACertPath),
		client.WithInsecure(true),
	)
	if err != nil {
		h.t.Fatal(err)
	}

	return &DirectoryClient{
		Model:    dsm3.NewModelClient(c),
		Reader:   dsr3.NewReaderClient(c),
		Writer:   dsw3.NewWriterClient(c),
		Importer: dsi3.NewImporterClient(c),
		Exporter: dse3.NewExporterClient(c),
	}
}

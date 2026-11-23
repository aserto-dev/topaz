package ds

import (
	"github.com/aserto-dev/topaz/api/client"
	"github.com/aserto-dev/topaz/api/client/internal/hosted"
	des "github.com/aserto-dev/topaz/api/directory/v3/exporter"
	dis "github.com/aserto-dev/topaz/api/directory/v3/importer"
	dms "github.com/aserto-dev/topaz/api/directory/v3/model"
	drs "github.com/aserto-dev/topaz/api/directory/v3/reader"
	dws "github.com/aserto-dev/topaz/api/directory/v3/writer"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto Directory APIs.
type Client struct {
	// Client for the directory reader service.
	Reader drs.ReaderClient

	// Client for the directory writer service.
	Writer dws.WriterClient

	// Client for the directory importer service.
	Importer dis.ImporterClient

	// Client for the directory exporter service.
	Exporter des.ExporterClient

	// Client for the directory model service.
	Model dms.ModelClient

	conns []*grpc.ClientConn
}

// New returns a new Client with the specified options.
func New(opts ...client.ConnectionOption) (*Client, error) {
	options, err := client.NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.Address == "" {
		options.Address = hosted.HostedDirectoryHostname + hosted.HostedDirectoryGRPCPort
	}

	conn, err := client.Connect(options)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		Reader:   drs.NewReaderClient(conn),
		Writer:   dws.NewWriterClient(conn),
		Importer: dis.NewImporterClient(conn),
		Exporter: des.NewExporterClient(conn),
		Model:    dms.NewModelClient(conn),
		conns:    []*grpc.ClientConn{conn},
	}, nil
}

// FromConnection returns a new Client using an existing connection.
func FromConnection(conn *grpc.ClientConn) *Client {
	return &Client{
		Reader:   drs.NewReaderClient(conn),
		Writer:   dws.NewWriterClient(conn),
		Importer: dis.NewImporterClient(conn),
		Exporter: des.NewExporterClient(conn),
		Model:    dms.NewModelClient(conn),
		conns:    []*grpc.ClientConn{conn},
	}
}

// Close closes the underlying connections.
func (c *Client) Close() error {
	var errs error

	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

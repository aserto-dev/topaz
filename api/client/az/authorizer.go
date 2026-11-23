package az

import (
	"github.com/aserto-dev/topaz/api/client"
	"google.golang.org/grpc"

	authz "github.com/aserto-dev/topaz/api/authorizer/v2"

	"github.com/pkg/errors"
)

// Client provides access to the Aserto authorization services.
type Client struct {
	authz.AuthorizerClient

	conn *grpc.ClientConn
}

// New creates a Client with the specified connection options.
func New(opts ...client.ConnectionOption) (*Client, error) {
	conn, err := client.NewConnection(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		AuthorizerClient: authz.NewAuthorizerClient(conn),
		conn:             conn,
	}, err
}

// FromConnection returns a new Client using an existing connection.
func FromConnection(conn *grpc.ClientConn) *Client {
	return &Client{
		AuthorizerClient: authz.NewAuthorizerClient(conn),
		conn:             conn,
	}
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn
}

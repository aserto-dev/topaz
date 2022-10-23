package clients

import (
	asertogoClient "github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/go-directory-cli/client"
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/google/uuid"
)

const localhostDirectory = "localhost:9292"

func NewDirectoryClient(c *cc.CommonCtx, address string) (*client.Client, error) {

	if address == "" {
		address = localhostDirectory
	}

	conn, err := asertogoClient.NewConnection(
		c.Context, asertogoClient.WithInsecure(true),
		asertogoClient.WithAddr(address),
		asertogoClient.WithSessionID(uuid.NewString()),
	)

	if err != nil {
		return nil, err
	}

	return client.New(conn, c.UI)
}

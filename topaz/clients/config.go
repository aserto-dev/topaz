package clients

import (
	"time"

	client "github.com/aserto-dev/topaz/api/client"
)

type Config interface {
	ClientConfig() *client.Config
	CommandTimeout() time.Duration
}

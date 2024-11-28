package clients

import (
	"time"

	client "github.com/aserto-dev/go-aserto"
)

type Config interface {
	ClientConfig() *client.Config
	CommandTimeout() time.Duration
}

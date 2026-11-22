package internal

import (
	"encoding/json"
	"hash/maphash"

	"github.com/aserto-dev/topaz/api/client"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type Connections struct {
	conns   map[uint64]*grpc.ClientConn
	seed    maphash.Seed
	Connect func(*client.Config) (*grpc.ClientConn, error)
}

func NewConnections() *Connections {
	return &Connections{
		conns: make(map[uint64]*grpc.ClientConn),
		seed:  maphash.MakeSeed(),
		Connect: func(cfg *client.Config) (*grpc.ClientConn, error) {
			return cfg.Connect()
		},
	}
}

func (cb *Connections) Get(cfg *client.Config) (*grpc.ClientConn, error) {
	bin, err := json.Marshal(cfg) //nolint:gosec
	if err != nil {
		return nil, err
	}

	hash := maphash.Bytes(cb.seed, bin)

	conn := cb.conns[hash]
	if conn == nil {
		conn, err = cb.Connect(cfg)
		if err != nil {
			return nil, err
		}

		cb.conns[hash] = conn
	}

	return conn, nil
}

func (cb *Connections) AsSlice() []*grpc.ClientConn {
	return lo.Values(cb.conns)
}

// ConnectCounter used for testing.
type ConnectCounter struct {
	Count int
}

func (cc *ConnectCounter) Connect(*client.Config) (*grpc.ClientConn, error) {
	cc.Count++
	return &grpc.ClientConn{}, nil
}

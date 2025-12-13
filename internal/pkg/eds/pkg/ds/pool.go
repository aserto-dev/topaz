package ds

import (
	"bytes"

	"github.com/aserto-dev/azm/mempool"
)

const maxRelationIdentifierSize = 384

var relidBufPool = mempool.NewPool[*bytes.Buffer](NewRelationIdentifierBuf)

func NewRelationIdentifierBuf() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, maxRelationIdentifierSize))
}

func RelationIdentifierBuffer() *bytes.Buffer {
	return relidBufPool.Get()
}

func ReturnRelationIdentifierBuffer(b *bytes.Buffer) {
	b.Reset()
	relidBufPool.Put(b)
}

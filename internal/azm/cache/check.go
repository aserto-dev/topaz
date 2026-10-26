package cache

import (
	dsr "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/pkg/pb"
	"github.com/aserto-dev/go-directory/pkg/prop"
	"github.com/aserto-dev/topaz/internal/azm/graph"
	"github.com/aserto-dev/topaz/internal/azm/mempool"
	"google.golang.org/protobuf/types/known/structpb"
)

// If true, use a shared memory pool for all requests.
// Othersise, each call gets its own pool.
const sharedPool = true

func (c *Cache) Check(req *dsr.CheckRequest, relReader graph.RelationReader) (*dsr.CheckResponse, error) {
	checker := graph.NewCheck(c.model.Load(), req, relReader, c.relationsPool())

	ctx := pb.NewStruct()

	var reason string

	ok, err := checker.Check()
	if err != nil {
		reason = err.Error()
	} else {
		reason = checker.Reason()
	}

	if reason != "" {
		ctx.Fields[prop.Reason] = structpb.NewStringValue(reason)
	}

	return &dsr.CheckResponse{Check: ok, Trace: checker.Trace(), Context: ctx}, nil
}

type graphSearch interface {
	Search() (*dsr.GetGraphResponse, error)
}

func (c *Cache) GetGraph(req *dsr.GetGraphRequest, relReader graph.RelationReader) (*dsr.GetGraphResponse, error) {
	var (
		search graphSearch
		err    error
	)

	if req.GetObjectId() == "" {
		search, err = graph.NewObjectSearch(c.model.Load(), req, relReader, c.relationsPool())
	} else {
		search, err = graph.NewSubjectSearch(c.model.Load(), req, relReader, c.relationsPool())
	}

	if err != nil {
		return nil, err
	}

	return search.Search()
}

func (c *Cache) relationsPool() *mempool.RelationsPool {
	if sharedPool {
		return c.relsPool
	}

	return mempool.NewRelationsPool()
}

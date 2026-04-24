package cache

import (
	"github.com/aserto-dev/topaz/api/directory/pkg/pb"
	"github.com/aserto-dev/topaz/api/directory/pkg/prop"
	dsr "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/graph"
	"github.com/aserto-dev/topaz/azm/mempool"
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
	Search() (*dsr.GraphResponse, error)
}

func (c *Cache) GetGraph(req *dsr.GraphRequest, relReader graph.RelationReader) (*dsr.GraphResponse, error) {
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

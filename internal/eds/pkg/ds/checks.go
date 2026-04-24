package ds

import (
	"context"
	"runtime"

	"github.com/aserto-dev/topaz/api/directory/pkg/prop"
	dsr3 "github.com/aserto-dev/topaz/api/directory/v4/reader"
	"github.com/aserto-dev/topaz/azm/cache"
	"github.com/aserto-dev/topaz/azm/jobpool"
	"github.com/aserto-dev/topaz/azm/safe"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/structpb"
)

type checks struct {
	*safe.SafeChecks
}

func Checks(i *dsr3.ChecksRequest) *checks {
	if i.GetDefault() == nil {
		i.Default = &dsr3.CheckRequest{}
	}

	return &checks{safe.Checks(i)}
}

func (i *checks) Validate(mc *cache.Cache) error {
	return nil
}

func (i *checks) Exec(ctx context.Context, tx *bolt.Tx, mc *cache.Cache) (*dsr3.ChecksResponse, error) {
	consumer := func(in *dsr3.CheckRequest) *dsr3.CheckResponse {
		check := Check(in)
		if err := check.Validate(mc); err != nil {
			return checkError(err)
		}

		resp, err := check.Exec(ctx, tx, mc)
		if err != nil {
			return checkError(err)
		}

		return resp
	}

	pool := jobpool.NewJobPool(len(i.Checks), runtime.GOMAXPROCS(0), consumer)
	pool.Start()

	resp := &dsr3.ChecksResponse{}

	for check := range i.CheckRequests() {
		if err := pool.Produce(check.CheckRequest); err != nil {
			return resp, err
		}
	}

	resp.Checks = pool.Wait()

	return resp, nil
}

func checkError(err error) *dsr3.CheckResponse {
	return &dsr3.CheckResponse{
		Check: false,
		Context: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				prop.Reason: structpb.NewStringValue(err.Error()),
			},
		},
	}
}

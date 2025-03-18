package common

import (
	"github.com/aserto-dev/topaz/pkg/cli/cc"
	"github.com/aserto-dev/topaz/pkg/cli/jsonx"
	"github.com/aserto-dev/topaz/pkg/cli/pb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Message[T any] interface {
	proto.Message
	*T
}

func Invoke[T any, M Message[T]](c *cc.CommonCtx, conn grpc.ClientConnInterface, method, request string) error {
	var req M
	if err := pb.UnmarshalRequest(request, req); err != nil {
		return err
	}

	var resp proto.Message
	if err := conn.Invoke(c.Context, method, req, resp); err != nil {
		return err
	}

	return jsonx.OutputJSONPB(c.StdOut(), resp)
}

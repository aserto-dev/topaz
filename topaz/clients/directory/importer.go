package directory

import (
	"bufio"
	"context"
	"errors"
	"io"

	dsc3 "github.com/aserto-dev/go-directory/aserto/directory/common/v3"
	dsi3 "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	"github.com/aserto-dev/topaz/topaz/jsonx"
	"golang.org/x/sync/errgroup"
)

func (c *Client) ImportFromFile(ctx context.Context, r io.Reader) error {
	errGrp, errGrpCtx := errgroup.WithContext(ctx)

	stream, err := c.Importer.Import(errGrpCtx)
	if err != nil {
		return err
	}

	errGrp.Go(c.recv(stream))

	errGrp.Go(c.sender(stream, r))

	return errGrp.Wait()
}

func (c *Client) recv(stream dsi3.Importer_ImportClient) func() error {
	return func() error {
		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}

			switch m := msg.GetMsg().(type) {
			case *dsi3.ImportResponse_Status:
				_ = m.Status

			case *dsi3.ImportResponse_Counter:
				_ = m.Counter

			default:
				_ = m
			}
		}
	}
}

func (c *Client) sender(stream dsi3.Importer_ImportClient, r io.Reader) func() error {
	return func() error {
		if err := c.importFromReader(stream, r); err != nil {
			return err
		}

		if err := stream.CloseSend(); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) importFromReader(stream dsi3.Importer_ImportClient, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		obj := &dsc3.Object{}
		rel := &dsc3.Relation{}

		if err := jsonx.Unmarshal(scanner.Bytes(), obj); err == nil {
			// fmt.Fprintf(os.Stderr, "obj:[%s:%s]\n", obj.GetType(), obj.GetId())

			if err := stream.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg: &dsi3.ImportRequest_Object{
					Object: obj,
				},
			}); err != nil {
				return err
			}

			continue
		}

		if err := jsonx.Unmarshal(scanner.Bytes(), rel); err == nil {
			// fmt.Fprintf(os.Stderr, "rel:[%s:%s#%s@%s:%s#%s]\n",
			// 	rel.GetObjectId(),
			// 	rel.GetObjectId(),
			// 	rel.GetRelation(),
			// 	rel.GetSubjectType(),
			// 	rel.GetSubjectId(),
			// 	rel.GetSubjectRelation(),
			// )

			if err := stream.Send(&dsi3.ImportRequest{
				OpCode: dsi3.Opcode_OPCODE_SET,
				Msg: &dsi3.ImportRequest_Relation{
					Relation: rel,
				},
			}); err != nil {
				return err
			}

			continue
		}
	}

	return scanner.Err()
}
